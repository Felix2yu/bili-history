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
    ps: int = Query(20, description="每页数量"),
):
    try:
        config = load_config()
        vmid = config.get("DedeUserID", "")
        if not vmid:
            return {"status": "error", "message": "未登录，无法获取点赞列表"}

        url = "https://api.bilibili.com/x/space/like/video"
        params = {"vmid": vmid, "pn": pn, "ps": ps}
        headers = get_headers()
        response = requests.get(url, params=params, headers=headers)
        data = response.json()
        if data.get("code") != 0:
            code = data.get("code")
            message = data.get("message", "未知错误")
            if code == -6:
                message = "Cookie 已过期，请重新登录"
            return {"status": "error", "message": message, "code": code}

        list_data = data.get("data", {}).get("list", [])
        total = data.get("data", {}).get("page", {}).get("count", 0)
        if not total:
            total = data.get("data", {}).get("total", 0)
        if not total:
            total = len(list_data)
        result = []
        for item in list_data:
            owner = item.get("owner", {})
            stat = item.get("stat", {})
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
                "pubdate": item.get("pubdate", 0),
                "view": stat.get("view", 0),
                "danmaku": stat.get("danmaku", 0),
                "like": stat.get("like", 0),
                "link": f"https://www.bilibili.com/video/{item.get('bvid', '')}",
            })
        return {"status": "success", "data": {"list": result, "total": total, "pn": pn, "ps": ps}}
    except Exception as e:
        return {"status": "error", "message": f"获取点赞列表失败: {str(e)}"}
