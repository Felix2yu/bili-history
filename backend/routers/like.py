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


@router.get("/list", summary="获取点赞视频列表")
async def get_like_list(
    pn: int = Query(1, description="页码"),
    ps: int = Query(50, description="每页数量"),
):
    try:
        config = load_config()
        vmid = config.get("DedeUserID", "")
        if not vmid:
            return {"status": "error", "message": "未登录，无法获取点赞列表"}

        all_results = []
        current_pn = 1
        max_pages = 50

        while current_pn <= max_pages:
            url = "https://api.bilibili.com/x/space/like/video"
            params = {"vmid": vmid, "pn": current_pn, "ps": ps}
            headers = get_headers()
            response = requests.get(url, params=params, headers=headers)
            data = response.json()
            if data.get("code") != 0:
                code = data.get("code")
                message = data.get("message", "未知错误")
                if code == -6:
                    message = "Cookie 已过期，请重新登录"
                if current_pn == 1:
                    return {"status": "error", "message": message, "code": code}
                break

            list_data = data.get("data", {}).get("list", [])
            if not list_data:
                break

            for item in list_data:
                owner = item.get("owner", {})
                stat = item.get("stat", {})
                all_results.append({
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
                    "pubdate": item.get("pubdate", 0),
                    "view": stat.get("view", 0),
                    "danmaku": stat.get("danmaku", 0),
                    "like": stat.get("like", 0),
                    "link": f"https://www.bilibili.com/video/{item.get('bvid', '')}",
                })

            if len(list_data) < ps:
                break
            current_pn += 1

        total = data.get("data", {}).get("page", {}).get("count", 0) if 'data' in dir() else 0
        if not total or total < len(all_results):
            total = len(all_results)

        return {"status": "success", "data": {"list": all_results, "total": total, "pn": 1, "ps": len(all_results)}}
    except Exception as e:
        return {"status": "error", "message": f"获取点赞列表失败: {str(e)}"}
