import os
import sqlite3
import time
import requests
from fastapi import APIRouter, Query
from scripts.utils import load_config, get_database_path

router = APIRouter()

DB_PATH = get_database_path("bilibili_likes.db")

CREATE_TABLE = """
CREATE TABLE IF NOT EXISTS liked_videos (
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
    pubdate INTEGER DEFAULT 0,
    view INTEGER DEFAULT 0,
    danmaku INTEGER DEFAULT 0,
    like_count INTEGER DEFAULT 0,
    link TEXT,
    fetch_time INTEGER NOT NULL,
    is_seen INTEGER DEFAULT 0
);
"""

CREATE_INDEXES = [
    "CREATE INDEX IF NOT EXISTS idx_liked_bvid ON liked_videos(bvid);",
    "CREATE INDEX IF NOT EXISTS idx_liked_pubdate ON liked_videos(pubdate);",
    "CREATE INDEX IF NOT EXISTS idx_liked_owner ON liked_videos(owner_name);",
    "CREATE INDEX IF NOT EXISTS idx_liked_tid ON liked_videos(tid);",
    "CREATE INDEX IF NOT EXISTS idx_liked_fetch_time ON liked_videos(fetch_time);",
    "CREATE INDEX IF NOT EXISTS idx_liked_is_seen ON liked_videos(is_seen);",
]

MIGRATE_IS_SEEN = "ALTER TABLE liked_videos ADD COLUMN is_seen INTEGER DEFAULT 0"


def get_db_connection():
    os.makedirs(os.path.dirname(DB_PATH), exist_ok=True)
    conn = sqlite3.connect(DB_PATH)
    conn.row_factory = sqlite3.Row
    cursor = conn.cursor()
    cursor.execute(CREATE_TABLE)
    try:
        cursor.execute(MIGRATE_IS_SEEN)
    except sqlite3.OperationalError:
        pass
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


def _upsert_video(cursor, v, now):
    cursor.execute(
        """INSERT INTO liked_videos
        (bvid, aid, title, pic, desc, duration, tid, tname,
         owner_name, owner_mid, owner_face, pubdate,
         view, danmaku, like_count, link, fetch_time, is_seen)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 1)
        ON CONFLICT(bvid) DO UPDATE SET
            title=excluded.title, pic=excluded.pic, desc=excluded.desc,
            duration=excluded.duration, tid=excluded.tid, tname=excluded.tname,
            owner_name=excluded.owner_name, owner_mid=excluded.owner_mid,
            owner_face=excluded.owner_face, pubdate=excluded.pubdate,
            view=excluded.view, danmaku=excluded.danmaku,
            like_count=excluded.like_count, link=excluded.link,
            fetch_time=excluded.fetch_time, is_seen=1""",
        (v["bvid"], v["aid"], v["title"], v["pic"], v["desc"],
         v["duration"], v["tid"], v["tname"], v["owner_name"],
         v["owner_mid"], v["owner_face"], v["pubdate"],
         v["view"], v["danmaku"], v["like"], v["link"], now),
    )


def _parse_api_item(item):
    owner = item.get("owner", {})
    stat = item.get("stat", {})
    return {
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
    }


@router.get("/list", summary="增量同步点赞视频")
async def get_like_list():
    try:
        config = load_config()
        vmid = config.get("DedeUserID", "")
        if not vmid:
            return {"status": "error", "message": "未登录，无法获取点赞列表"}

        conn = get_db_connection()
        try:
            cursor = conn.cursor()
            cursor.execute("UPDATE liked_videos SET is_seen = 0")
            conn.commit()

            now = int(time.time())
            current_pn = 1
            ps = 50
            max_pages = 50
            fetched_count = 0
            new_count = 0
            updated_count = 0

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

                stop_early = False
                for item in list_data:
                    bvid = item.get("bvid")
                    if not bvid:
                        continue

                    cursor.execute("SELECT is_seen FROM liked_videos WHERE bvid = ?", (bvid,))
                    row = cursor.fetchone()
                    if row and row["is_seen"] == 1:
                        stop_early = True
                        break

                    v = _parse_api_item(item)
                    old_fetch = 0
                    if row:
                        cursor.execute("SELECT fetch_time FROM liked_videos WHERE bvid = ?", (bvid,))
                        old_row = cursor.fetchone()
                        if old_row:
                            old_fetch = old_row["fetch_time"]
                        updated_count += 1
                    else:
                        new_count += 1

                    _upsert_video(cursor, v, now)
                    fetched_count += 1

                if stop_early:
                    break
                if len(list_data) < ps:
                    break
                current_pn += 1

            conn.commit()

            cursor.execute("SELECT COUNT(*) FROM liked_videos")
            total = cursor.fetchone()[0]

            return {
                "status": "success",
                "data": {
                    "total": total,
                    "fetched": fetched_count,
                    "new": new_count,
                    "updated": updated_count,
                    "pages": current_pn,
                },
            }
        finally:
            conn.close()
    except Exception as e:
        return {"status": "error", "message": f"获取点赞列表失败: {str(e)}"}


@router.get("/local", summary="从本地数据库获取点赞视频")
async def get_like_local(
    page: int = Query(1, ge=1, description="页码"),
    size: int = Query(50, ge=1, le=500, description="每页数量"),
    sort: str = Query("pubdate", description="排序字段: pubdate/fetch_time/duration/owner_name"),
    order: str = Query("desc", description="排序方向: asc/desc"),
):
    try:
        allowed_sorts = {"pubdate", "fetch_time", "duration", "owner_name", "view", "like_count"}
        if sort not in allowed_sorts:
            sort = "pubdate"
        if order not in ("asc", "desc"):
            order = "desc"

        conn = get_db_connection()
        cursor = conn.cursor()
        cursor.execute("SELECT COUNT(*) FROM liked_videos")
        total = cursor.fetchone()[0]

        offset = (page - 1) * size
        cursor.execute(
            f"SELECT * FROM liked_videos ORDER BY {sort} {order} LIMIT ? OFFSET ?",
            (size, offset),
        )
        rows = [dict(r) for r in cursor.fetchall()]
        has_more = offset + size < total

        conn.close()
        return {
            "status": "success",
            "data": {
                "list": rows,
                "total": total,
                "page": page,
                "size": size,
                "has_more": has_more,
            },
        }
    except Exception as e:
        return {"status": "error", "message": f"获取本地点赞数据失败: {str(e)}"}
