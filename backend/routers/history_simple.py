import os
import sqlite3
import time
import requests
from fastapi import APIRouter, Query
from scripts.utils import load_config, get_database_path

router = APIRouter()

DB_PATH = get_database_path("bilibili_history_simple.db")

CREATE_TABLE = """
CREATE TABLE IF NOT EXISTS history_videos (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    bvid TEXT NOT NULL,
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
    view_at INTEGER DEFAULT 0,
    progress INTEGER DEFAULT 0,
    view INTEGER DEFAULT 0,
    danmaku INTEGER DEFAULT 0,
    link TEXT,
    fetch_time INTEGER NOT NULL,
    UNIQUE(bvid, view_at)
);
"""

CREATE_INDEXES = [
    "CREATE INDEX IF NOT EXISTS idx_hist_bvid ON history_videos(bvid);",
    "CREATE INDEX IF NOT EXISTS idx_hist_view_at ON history_videos(view_at);",
    "CREATE INDEX IF NOT EXISTS idx_hist_owner ON history_videos(owner_name);",
    "CREATE INDEX IF NOT EXISTS idx_hist_tid ON history_videos(tid);",
    "CREATE INDEX IF NOT EXISTS idx_hist_fetch_time ON history_videos(fetch_time);",
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


def save_history_videos(conn, videos):
    cursor = conn.cursor()
    now = int(time.time())
    for v in videos:
        cursor.execute(
            """INSERT INTO history_videos
            (bvid, aid, title, pic, desc, duration, tid, tname,
             owner_name, owner_mid, owner_face, view_at, progress,
             view, danmaku, link, fetch_time)
            VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
            ON CONFLICT(bvid, view_at) DO UPDATE SET
                title=excluded.title, pic=excluded.pic, desc=excluded.desc,
                duration=excluded.duration, tid=excluded.tid, tname=excluded.tname,
                owner_name=excluded.owner_name, owner_mid=excluded.owner_mid,
                owner_face=excluded.owner_face, progress=excluded.progress,
                view=excluded.view, danmaku=excluded.danmaku, link=excluded.link,
                fetch_time=excluded.fetch_time""",
            (v["bvid"], v["aid"], v["title"], v["pic"], v["desc"],
             v["duration"], v["tid"], v["tname"], v["owner_name"],
             v["owner_mid"], v["owner_face"], v["view_at"], v["progress"],
             v["view"], v["danmaku"], v["link"], now),
        )
    conn.commit()


@router.get("/list", summary="获取观看历史记录")
async def get_history_list():
    try:
        config = load_config()
        sessdata = config.get("SESSDATA", "")
        if not sessdata:
            return {"status": "error", "message": "未登录，无法获取历史记录"}

        all_results = []
        max_views = 2000
        view_at = 0

        while len(all_results) < max_views:
            url = "https://api.bilibili.com/x/web-interface/history/cursor"
            params = {"view_at": view_at, "ps": 20}
            headers = get_headers()
            response = requests.get(url, params=params, headers=headers)
            data = response.json()
            if data.get("code") != 0:
                code = data.get("code")
                message = data.get("message", "未知错误")
                if code == -6:
                    message = "Cookie 已过期，请重新登录"
                if not all_results:
                    return {"status": "error", "message": message, "code": code}
                break

            list_data = data.get("data", {}).get("list", [])
            if not list_data:
                break

            for item in list_data:
                upper = item.get("upper", {})
                all_results.append({
                    "aid": item.get("aid"),
                    "bvid": item.get("bvid"),
                    "title": item.get("title"),
                    "pic": item.get("pic"),
                    "desc": item.get("desc", ""),
                    "duration": item.get("duration", 0),
                    "tid": item.get("tid", 0),
                    "tname": item.get("tname", ""),
                    "owner_name": upper.get("name", ""),
                    "owner_mid": upper.get("mid", 0),
                    "owner_face": upper.get("face", ""),
                    "view_at": item.get("view_at", 0),
                    "progress": item.get("progress", 0),
                    "view": item.get("stat", {}).get("view", 0) if item.get("stat") else 0,
                    "danmaku": item.get("stat", {}).get("danmaku", 0) if item.get("stat") else 0,
                    "link": f"https://www.bilibili.com/video/{item.get('bvid', '')}",
                })

            cursor = data.get("data", {}).get("cursor", {})
            if not cursor.get("is_end", True):
                view_at = cursor.get("view_at", 0)
            else:
                break

        if all_results:
            conn = get_db_connection()
            try:
                save_history_videos(conn, all_results)
            finally:
                conn.close()

        return {"status": "success", "data": {"list": all_results, "total": len(all_results)}}
    except Exception as e:
        return {"status": "error", "message": f"获取历史记录失败: {str(e)}"}


@router.get("/local", summary="从本地数据库获取历史记录")
async def get_history_local(
    page: int = Query(1, description="页码"),
    size: int = Query(50, description="每页数量"),
):
    try:
        conn = get_db_connection()
        cursor = conn.cursor()
        cursor.execute("SELECT COUNT(*) FROM history_videos")
        total = cursor.fetchone()[0]
        offset = (page - 1) * size
        cursor.execute(
            "SELECT * FROM history_videos ORDER BY view_at DESC LIMIT ? OFFSET ?",
            (size, offset),
        )
        rows = [dict(r) for r in cursor.fetchall()]
        conn.close()
        return {"status": "success", "data": {"list": rows, "total": total, "page": page, "size": size}}
    except Exception as e:
        return {"status": "error", "message": f"获取本地历史记录失败: {str(e)}"}
