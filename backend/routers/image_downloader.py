import os
import sqlite3
import hashlib
from typing import Optional
from urllib.parse import urlparse

from fastapi import APIRouter, BackgroundTasks, HTTPException, Query
from fastapi.responses import FileResponse
import requests as http_requests

from scripts.image_downloader import ImageDownloader
from scripts.utils import get_output_path

router = APIRouter()
downloader = ImageDownloader()

def get_history_db():
    """获取历史记录数据库连接"""
    db_path = get_output_path('bilibili_history.db')
    return sqlite3.connect(db_path)

@router.post("/start", summary="开始下载图片")
async def start_download(
    background_tasks: BackgroundTasks,
    year: Optional[int] = None,
    use_sessdata: bool = True
):
    """开始下载图片

    Args:
        year: 指定年份，不指定则下载所有年份
        use_sessdata: 是否在下载图片时使用SESSDATA（对于公开内容如视频封面和头像，可以不使用SESSDATA）
    """
    # 包装下载函数和状态更新
    def download_with_status_update(year=None, use_sessdata=True):
        try:
            # 执行下载
            downloader.start_download(year, use_sessdata)
        except Exception as e:
            print(f"下载过程发生错误: {str(e)}")
        finally:
            # 确保无论下载成功还是失败，状态都会被设置为已完成
            print("\n=== 下载任务完成，更新状态 ===")
            downloader.is_downloading = False
            print("下载状态已设置为已完成")

    # 在后台任务中执行包装函数
    background_tasks.add_task(download_with_status_update, year, use_sessdata)

    return {
        "status": "success",
        "message": f"开始下载{'所有年份' if year is None else f'{year}年'}的图片"
    }

@router.post("/stop", summary="停止下载图片")
async def stop_download():
    """停止当前下载任务

    Returns:
        dict: 包含停止状态和当前下载统计的响应
    """
    try:
        result = downloader.stop_download()
        return result
    except Exception as e:
        return {
            "status": "error",
            "message": f"停止下载失败: {str(e)}"
        }

@router.get("/status", summary="获取下载状态")
async def get_status():
    """获取下载状态"""
    stats = downloader.get_download_stats()

    return {
        "status": "success",
        "data": stats
    }

@router.post("/clear", summary="清空所有图片")
async def clear_images():
    """清空所有图片和下载状态"""
    try:
        success = downloader.clear_all_images()
        if success:
            return {
                "status": "success",
                "message": "已清空所有图片和下载状态",
                "data": {
                    "cleared_paths": [
                        "output/images/covers",
                        "output/images/avatars",
                        "output/images/orphaned_covers",
                        "output/images/orphaned_avatars"
                    ],
                    "status_file": "output/download_status.json"
                }
            }
        else:
            return {
                "status": "error",
                "message": "清空图片失败，请查看日志了解详细信息"
            }
    except Exception as e:
        return {
            "status": "error",
            "message": f"清空图片时发生错误: {str(e)}"
        }

@router.get("/local/{image_type}/{file_hash}", summary="获取本地图片")
async def get_local_image(image_type: str, file_hash: str):
    """获取本地图片

    Args:
        image_type: 图片类型 (covers 或 avatars)
        file_hash: 图片文件的哈希值

    Returns:
        FileResponse: 图片文件响应
    """

    # 验证图片类型
    if image_type not in ('covers', 'avatars'):
        raise HTTPException(
            status_code=400,
            detail=f"无效的图片类型: {image_type}"
        )

    try:
        # 构建图片路径
        base_path = get_output_path('images')
        type_path = os.path.join(base_path, image_type)
        sub_dir = file_hash[:2]  # 使用哈希的前两位作为子目录

        # 获取所有年份目录
        years = []
        if os.path.exists(type_path):
            for item in os.listdir(type_path):
                if item.isdigit():
                    years.append(item)

        # 按年份倒序搜索图片
        for year in sorted(years, reverse=True):
            year_path = os.path.join(type_path, year)
            img_dir = os.path.join(year_path, sub_dir)

            if not os.path.exists(img_dir):
                continue

            # 查找所有可能的图片文件扩展名
            for ext in ('.jpg', '.jpeg', '.png', '.webp', '.gif'):
                img_path = os.path.join(img_dir, f"{file_hash}{ext}")
                if os.path.exists(img_path):
                    print(f"找到图片文件: {img_path}")
                    return FileResponse(
                        img_path,
                        media_type=f"image/{ext[1:]}" if ext != '.jpg' else "image/jpeg"
                    )

        # 如果在年份目录中没有找到，尝试在根目录中查找
        img_dir = os.path.join(type_path, sub_dir)
        if os.path.exists(img_dir):
            for ext in ('.jpg', '.jpeg', '.png', '.webp', '.gif'):
                img_path = os.path.join(img_dir, f"{file_hash}{ext}")
                if os.path.exists(img_path):
                    print(f"找到图片文件: {img_path}")
                    return FileResponse(
                        img_path,
                        media_type=f"image/{ext[1:]}" if ext != '.jpg' else "image/jpeg"
                    )

        # 如果没有找到任何匹配的文件
        raise HTTPException(
            status_code=404,
            detail=f"图片不存在: {file_hash}"
        )

    except Exception as e:
        if isinstance(e, HTTPException):
            raise e
        print(f"获取本地图片时出错: {str(e)}")
        raise HTTPException(
            status_code=500,
            detail=f"获取图片失败: {str(e)}"
        )


@router.get("/proxy", summary="代理获取并缓存图片")
async def proxy_image(url: str = Query(..., description="图片URL")):
    """代理获取图片并缓存到本地

    接收任意图片URL，下载并缓存到 output/images/proxy/ 目录。
    已缓存的图片直接返回，不重复下载。
    """
    if not url:
        raise HTTPException(status_code=400, detail="URL不能为空")

    # 升级 http 为 https
    if url.startswith("http://"):
        url = url.replace("http://", "https://", 1)

    # 计算 URL 的 MD5 哈希作为文件名
    url_hash = hashlib.md5(url.encode()).hexdigest()
    sub_dir = url_hash[:2]

    # 缓存目录
    cache_base = get_output_path("images/proxy")
    cache_dir = os.path.join(cache_base, sub_dir)
    cache_file_base = os.path.join(cache_dir, url_hash)

    # 检查缓存是否存在
    for ext in ('.jpg', '.jpeg', '.png', '.webp', '.gif'):
        cached = cache_file_base + ext
        if os.path.exists(cached):
            media_type = "image/jpeg" if ext in ('.jpg', '.jpeg') else f"image/{ext[1:]}"
            return FileResponse(cached, media_type=media_type)

    # 缓存不存在，下载图片
    try:
        resp = http_requests.get(
            url,
            headers={
                "User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
                "Referer": "https://www.bilibili.com/",
            },
            timeout=15,
            stream=True,
        )
        resp.raise_for_status()

        # 从 Content-Type 推断扩展名
        content_type = resp.headers.get("Content-Type", "image/jpeg")
        ext_map = {
            "image/jpeg": ".jpg",
            "image/png": ".png",
            "image/webp": ".webp",
            "image/gif": ".gif",
        }
        ext = ext_map.get(content_type.split(";")[0].strip(), ".jpg")

        # 保存到缓存
        os.makedirs(cache_dir, exist_ok=True)
        cache_path = cache_file_base + ext
        with open(cache_path, "wb") as f:
            for chunk in resp.iter_content(8192):
                f.write(chunk)

        return FileResponse(cache_path, media_type=content_type)

    except http_requests.RequestException:
        raise HTTPException(status_code=502, detail="图片下载失败")