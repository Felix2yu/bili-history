import hashlib
import json
import os
import sqlite3
import time
from datetime import datetime
from typing import Dict, List, Optional, Tuple

from scripts.utils import get_database_path, get_output_path

DB_PATH = get_database_path("sync_state.db")

CREATE_TABLE = """
CREATE TABLE IF NOT EXISTS file_sync_state (
    file_path TEXT PRIMARY KEY,
    file_hash TEXT NOT NULL,
    record_count INTEGER DEFAULT 0,
    last_sync_time INTEGER NOT NULL,
    last_modified_time INTEGER NOT NULL
);
"""

CREATE_INDEXES = [
    "CREATE INDEX IF NOT EXISTS idx_sync_hash ON file_sync_state(file_hash);",
    "CREATE INDEX IF NOT EXISTS idx_sync_time ON file_sync_state(last_sync_time);",
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


def calculate_file_hash(file_path: str) -> str:
    """计算文件 MD5 哈希"""
    hash_md5 = hashlib.md5()
    try:
        with open(file_path, "rb") as f:
            for chunk in iter(lambda: f.read(4096), b""):
                hash_md5.update(chunk)
        return hash_md5.hexdigest()
    except Exception:
        return ""


def get_file_mtime(file_path: str) -> int:
    """获取文件修改时间"""
    try:
        return int(os.path.getmtime(file_path))
    except Exception:
        return 0


def get_file_record_count(file_path: str) -> int:
    """获取 JSON 文件中的记录数"""
    try:
        for encoding in ['utf-8', 'gbk', 'utf-8-sig']:
            try:
                with open(file_path, 'r', encoding=encoding) as f:
                    data = json.load(f)
                    return len(data) if isinstance(data, list) else 0
            except (UnicodeDecodeError, json.JSONDecodeError):
                continue
        return 0
    except Exception:
        return 0


def get_sync_state(file_path: str) -> Optional[dict]:
    """获取单个文件的同步状态"""
    conn = get_db_connection()
    try:
        cursor = conn.cursor()
        cursor.execute("SELECT * FROM file_sync_state WHERE file_path = ?", (file_path,))
        row = cursor.fetchone()
        return dict(row) if row else None
    finally:
        conn.close()


def update_sync_state(file_path: str, file_hash: str, record_count: int):
    """更新文件同步状态"""
    conn = get_db_connection()
    try:
        cursor = conn.cursor()
        now = int(time.time())
        mtime = get_file_mtime(file_path)
        cursor.execute(
            """INSERT INTO file_sync_state (file_path, file_hash, record_count, last_sync_time, last_modified_time)
            VALUES (?, ?, ?, ?, ?)
            ON CONFLICT(file_path) DO UPDATE SET
                file_hash=excluded.file_hash,
                record_count=excluded.record_count,
                last_sync_time=excluded.last_sync_time,
                last_modified_time=excluded.last_modified_time""",
            (file_path, file_hash, record_count, now, mtime),
        )
        conn.commit()
    finally:
        conn.close()


def is_file_changed(file_path: str) -> bool:
    """检查文件是否已更改"""
    state = get_sync_state(file_path)
    if state is None:
        return True

    current_hash = calculate_file_hash(file_path)
    current_mtime = get_file_mtime(file_path)

    if current_hash != state['file_hash']:
        return True
    if current_mtime != state['last_modified_time']:
        return True

    return False


def get_all_sync_states() -> List[dict]:
    """获取所有文件的同步状态"""
    conn = get_db_connection()
    try:
        cursor = conn.cursor()
        cursor.execute("SELECT * FROM file_sync_state ORDER BY last_sync_time DESC")
        return [dict(row) for row in cursor.fetchall()]
    finally:
        conn.close()


def get_changed_files(json_root: str) -> List[str]:
    """获取所有已更改的 JSON 文件"""
    changed_files = []
    if not os.path.exists(json_root):
        return changed_files

    for year_dir in sorted(os.listdir(json_root)):
        year_path = os.path.join(json_root, year_dir)
        if not os.path.isdir(year_path):
            continue
        for month_dir in sorted(os.listdir(year_path)):
            month_path = os.path.join(year_path, month_dir)
            if not os.path.isdir(month_path):
                continue
            for day_file in sorted(os.listdir(month_path)):
                if not day_file.endswith('.json'):
                    continue
                file_path = os.path.join(month_path, day_file)
                if is_file_changed(file_path):
                    changed_files.append(file_path)

    return changed_files


def get_sync_summary() -> dict:
    """获取同步摘要"""
    states = get_all_sync_states()
    total_files = len(states)
    total_records = sum(s['record_count'] for s in states)
    total_size = 0
    for s in states:
        try:
            total_size += os.path.getsize(s['file_path'])
        except Exception:
            pass

    last_sync = max((s['last_sync_time'] for s in states), default=0)

    return {
        "total_files": total_files,
        "total_records": total_records,
        "total_size_mb": round(total_size / (1024 * 1024), 2),
        "last_sync_time": last_sync,
        "last_sync_datetime": datetime.fromtimestamp(last_sync).isoformat() if last_sync else None,
    }
