import requests
from fastapi import APIRouter, Query
from scripts.utils import load_config

router = APIRouter()


def get_headers():
    config = load_config()
    sessdata = config.get("SESSDATA", "")
    headers = {
        "User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
        "Referer": "https://www.bilibili.com/",
    }
    cookies = []
    if sessdata:
        cookies.append(f"SESSDATA={sessdata}")
    bili_jct = config.get("bili_jct", "")
    if bili_jct:
        cookies.append(f"bili_jct={bili_jct}")
    dede_user_id = config.get("DedeUserID", "")
    if dede_user_id:
        cookies.append(f"DedeUserID={dede_user_id}")
    if cookies:
        headers["Cookie"] = "; ".join(cookies)
    return headers


@router.get("/list", summary="获取稍后再看列表")
async def get_watch_later_list():
    try:
        url = "https://api.bilibili.com/x/v2/history/toview"
        headers = get_headers()
        response = requests.get(url, headers=headers)
        data = response.json()
        if data.get("code") != 0:
            code = data.get("code")
            message = data.get("message", "未知错误")
            if code == -6:
                message = "Cookie 已过期，请重新登录"
            return {"status": "error", "message": message, "code": code}
        list_data = data.get("data", {}).get("list", [])
        result = []
        for item in list_data:
            owner = item.get("owner", {})
            result.append({
                "aid": item.get("aid"),
                "bvid": item.get("bvid"),
                "title": item.get("title"),
                "pic": item.get("pic"),
                "desc": item.get("desc", ""),
                "duration": item.get("duration", 0),
                "tid": item.get("tid", 0),
                "tname": item.get("tname", ""),
                "owner_name": owner.get("name", ""),
                "owner_mid": owner.get("mid", 0),
                "owner_face": owner.get("face", ""),
                "add_at": item.get("add_at", 0),
                "pubdate": item.get("pubdate", 0),
                "view": item.get("stat", {}).get("view", 0),
                "danmaku": item.get("stat", {}).get("danmaku", 0),
                "link": f"https://www.bilibili.com/video/{item.get('bvid', '')}",
            })
        return {"status": "success", "data": {"list": result, "total": len(result)}}
    except Exception as e:
        return {"status": "error", "message": f"获取稍后再看列表失败: {str(e)}"}


@router.delete("/{bvid}", summary="从稍后再看中移除视频")
async def remove_from_watch_later(bvid: str):
    try:
        config = load_config()
        bili_jct = config.get("bili_jct", "")
        url = "https://api.bilibili.com/x/v2/history/toview/del"
        headers = get_headers()
        data = {"bvid": bvid}
        response = requests.post(url, json=data, headers=headers)
        result = response.json()
        if result.get("code") != 0:
            return {"status": "error", "message": result.get("message", "移除失败"), "code": result.get("code")}
        return {"status": "success", "message": "已从稍后再看中移除"}
    except Exception as e:
        return {"status": "error", "message": f"移除失败: {str(e)}"}
