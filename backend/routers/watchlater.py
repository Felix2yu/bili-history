import os
import sqlite3
import time
import requests
from fastapi import APIRouter, Query
from scripts.utils import load_config, get_database_path

router = APIRouter()

DB_PATH = get_database_path("bilibili_watchlater.db")

CREATE_TABLE = """
CREATE TABLE IF NOT EXISTS watchlater_videos (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    bvid TEXT NOT NULL UNIQUE,
    aid INTEGER,
    title TEXT NOT NULL,
    pic TEXT,
    desc TEXT,
    duration INTEGER DEFAULT 0,
    tid INTEGER DEFAULT 0,
    tname TEXT,
    owner_name TEXT,
    owner_mid INTEGER DEFAULT 0,
    owner_face TEXT,
    add_at INTEGER DEFAULT 0,
    pubdate INTEGER DEFAULT 0,
    view INTEGER DEFAULT 0,
    danmaku INTEGER DEFAULT 0,
    link TEXT,
    fetch_time INTEGER NOT NULL
);
"""

CREATE_INDEXES = [
    "CREATE INDEX IF NOT EXISTS idx_wl_bvid ON watchlater_videos(bvid);",
    "CREATE INDEX IF NOT EXISTS idx_wl_add_at ON watchlater_videos(add_at);",
    "CREATE INDEX IF NOT EXISTS idx_wl_owner ON watchlater_videos(owner_name);",
    "CREATE INDEX IF NOT EXISTS idx_wl_tid ON watchlater_videos(tid);",
    "CREATE INDEX IF NOT EXISTS idx_wl_fetch_time ON watchlater_videos(fetch_time);",
]


def get_db_connection():
    os.makedirs(os.path.dirname(DB_PATH), exist_ok=True)
    conn = sqlite3.connect(DB_PATH)
    conn.row_factory = sqlite3.Row
    cursor = conn.cursor()
    cursor.execute(CREATE_TABLE)
    for sql in CREATE_INDEXES:
        cursor.execute(sql)
    conn.commit()
    return conn


def get_headers():
    config = load_config()
    sessdata = config.get("SESSDATA", "")
    headers = {
        "User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
        "Referer": "https://www.bilibili.com/",
        "Origin": "https://www.bilibili.com",
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


def save_watchlater_videos(conn, videos):
    cursor = conn.cursor()
    now = int(time.time())
    for v in videos:
        cursor.execute(
            """INSERT INTO watchlater_videos
            (bvid, aid, title, pic, desc, duration, tid, tname,
             owner_name, owner_mid, owner_face, add_at, pubdate,
             view, danmaku, link, fetch_time)
            VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
            ON CONFLICT(bvid) DO UPDATE SET
                title=excluded.title, pic=excluded.pic, desc=excluded.desc,
                duration=excluded.duration, tid=excluded.tid, tname=excluded.tname,
                owner_name=excluded.owner_name, owner_mid=excluded.owner_mid,
                owner_face=excluded.owner_face, add_at=excluded.add_at,
                pubdate=excluded.pubdate, view=excluded.view,
                danmaku=excluded.danmaku, link=excluded.link,
                fetch_time=excluded.fetch_time""",
            (v["bvid"], v["aid"], v["title"], v["pic"], v["desc"],
             v["duration"], v["tid"], v["tname"], v["owner_name"],
             v["owner_mid"], v["owner_face"], v["add_at"], v["pubdate"],
             v["view"], v["danmaku"], v["link"], now),
        )
    conn.commit()


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
                "add_at": item.get("add_at", 0),
                "pubdate": item.get("pubdate", 0),
                "view": stat.get("view", 0),
                "danmaku": stat.get("danmaku", 0),
                "link": f"https://www.bilibili.com/video/{item.get('bvid', '')}",
            })

        if result:
            conn = get_db_connection()
            try:
                save_watchlater_videos(conn, result)
            finally:
                conn.close()

        return {"status": "success", "data": {"list": result, "total": len(result)}}
    except Exception as e:
        return {"status": "error", "message": f"获取稍后再看列表失败: {str(e)}"}


@router.get("/local", summary="从本地数据库获取稍后再看列表")
async def get_watch_later_local(
    page: int = Query(1, description="页码"),
    size: int = Query(50, description="每页数量"),
):
    try:
        conn = get_db_connection()
        cursor = conn.cursor()
        cursor.execute("SELECT COUNT(*) FROM watchlater_videos")
        total = cursor.fetchone()[0]
        offset = (page - 1) * size
        cursor.execute(
            "SELECT * FROM watchlater_videos ORDER BY add_at DESC LIMIT ? OFFSET ?",
            (size, offset),
        )
        rows = [dict(r) for r in cursor.fetchall()]
        conn.close()
        return {"status": "success", "data": {"list": rows, "total": total, "page": page, "size": size}}
    except Exception as e:
        return {"status": "error", "message": f"获取本地稍后再看数据失败: {str(e)}"}


@router.delete("/{bvid}", summary="从稍后再看中移除视频")
async def remove_from_watch_later(bvid: str, viewed: int = 0):
    try:
        # 获取配置
        config = load_config()
        bili_jct = config.get("bili_jct", "")

        if not bili_jct or bili_jct.startswith("你的"):
            return {"status": "error", "message": "缺少CSRF Token (bili_jct)，请先使用QR码登录并确保已正确获取bili_jct"}

        url = "https://api.bilibili.com/x/v2/history/toview/del"

        # 使用 Session 保持 cookie 一致性
        session = requests.Session()
        session.headers.update({
            "User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
            "Referer": "https://www.bilibili.com/",
            "Origin": "https://www.bilibili.com",
        })
        # 设置 cookie
        sessdata = config.get("SESSDATA", "")
        dede_user_id = config.get("DedeUserID", "")
        if sessdata:
            session.cookies.set("SESSDATA", sessdata, domain=".bilibili.com")
        if bili_jct:
            session.cookies.set("bili_jct", bili_jct, domain=".bilibili.com")
        if dede_user_id:
            session.cookies.set("DedeUserID", dede_user_id, domain=".bilibili.com")

        # 构造请求体
        payload = {"bvid": bvid, "csrf": bili_jct}
        if viewed:
            payload["viewed"] = viewed

        response = session.post(url, data=payload)
        result = response.json()

        from loguru import logger
        logger.info(f"[watchlater-del] bvid={bvid}, viewed={viewed}, response_code={result.get('code')}, message={result.get('message')}")

        if result.get("code") != 0:
            return {"status": "error", "message": result.get("message", "移除失败"), "code": result.get("code")}

        conn = get_db_connection()
        try:
            conn.execute("DELETE FROM watchlater_videos WHERE bvid = ?", (bvid,))
            conn.commit()
        finally:
            conn.close()

        return {"status": "success", "message": "已从稍后再看中移除"}
    except Exception as e:
        return {"status": "error", "message": f"移除失败: {str(e)}"}
