import json
import logging
import os
import sqlite3
import time
from concurrent.futures import ThreadPoolExecutor, as_completed
from datetime import datetime
from typing import List, Optional, Tuple

from scripts.utils import load_config, get_output_path
from scripts.sync_state import (
    calculate_file_hash,
    get_file_record_count,
    is_file_changed,
    update_sync_state,
    get_changed_files,
    get_sync_summary,
)

logger = logging.getLogger(__name__)
config = load_config()


def get_db_connection(db_path: str):
    """获取数据库连接"""
    os.makedirs(os.path.dirname(db_path), exist_ok=True)
    conn = sqlite3.connect(db_path)
    conn.row_factory = sqlite3.Row
    return conn


def ensure_history_table(conn, year: int):
    """确保历史记录表存在"""
    table_name = f"bilibili_history_{year}"
    cursor = conn.cursor()
    cursor.execute(f"""
        CREATE TABLE IF NOT EXISTS {table_name} (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            bvid TEXT NOT NULL,
            aid INTEGER,
            title TEXT,
            desc TEXT,
            pic TEXT,
            duration INTEGER,
            owner_name TEXT,
            owner_mid INTEGER,
            tag_name TEXT,
            tid INTEGER,
            view_at INTEGER NOT NULL,
            progress INTEGER,
            business TEXT,
            view INTEGER,
            danmaku INTEGER,
            coin INTEGER,
            favorite INTEGER,
            like INTEGER,
            reply INTEGER,
            share INTEGER,
            PRIMARY KEY (bvid, view_at)
        )
    """)
    cursor.execute(f"CREATE INDEX IF NOT EXISTS idx_{table_name}_bvid ON {table_name}(bvid)")
    cursor.execute(f"CREATE INDEX IF NOT EXISTS idx_{table_name}_view_at ON {table_name}(view_at)")
    cursor.execute(f"CREATE INDEX IF NOT EXISTS idx_{table_name}_owner ON {table_name}(owner_name)")
    conn.commit()


def import_single_json_file(file_path: str, db_path: str, force: bool = False) -> Tuple[int, int]:
    """导入单个 JSON 文件到数据库"""
    if not force and not is_file_changed(file_path):
        logger.debug(f"跳过未更改的文件: {file_path}")
        return 0, 0

    try:
        data = None
        for encoding in ['utf-8', 'gbk', 'utf-8-sig']:
            try:
                with open(file_path, 'r', encoding=encoding) as f:
                    data = json.load(f)
                break
            except (UnicodeDecodeError, json.JSONDecodeError):
                continue

        if data is None or not isinstance(data, list):
            logger.error(f"无法读取文件 {file_path}")
            return 0, 0

        conn = get_db_connection(db_path)
        try:
            inserted = 0
            skipped = 0
            cursor = conn.cursor()

            for item in data:
                view_at = item.get('view_at', 0)
                if view_at == 0:
                    continue

                history = item.get('history', {})
                bvid = history.get('bvid', '')
                if not bvid:
                    continue

                year = datetime.fromtimestamp(view_at).year
                table_name = f"bilibili_history_{year}"
                ensure_history_table(conn, year)

                cursor.execute(
                    f"SELECT 1 FROM {table_name} WHERE bvid = ? AND view_at = ?",
                    (bvid, view_at),
                )
                if cursor.fetchone():
                    skipped += 1
                    continue

                cursor.execute(f"""
                    INSERT INTO {table_name}
                    (bvid, aid, title, desc, pic, duration, owner_name, owner_mid,
                     tag_name, tid, view_at, progress, business, view, danmaku,
                     coin, favorite, like, reply, share)
                    VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
                """, (
                    bvid,
                    history.get('aid', 0),
                    item.get('title', ''),
                    item.get('desc', ''),
                    item.get('pic', ''),
                    item.get('duration', 0),
                    history.get('author', ''),
                    history.get('mid', 0),
                    item.get('tag_name', ''),
                    item.get('tid', 0),
                    view_at,
                    item.get('progress', 0),
                    history.get('business', ''),
                    item.get('stat', {}).get('view', 0) if item.get('stat') else 0,
                    item.get('stat', {}).get('danmaku', 0) if item.get('stat') else 0,
                    item.get('stat', {}).get('coin', 0) if item.get('stat') else 0,
                    item.get('stat', {}).get('favorite', 0) if item.get('stat') else 0,
                    item.get('stat', {}).get('like', 0) if item.get('stat') else 0,
                    item.get('stat', {}).get('reply', 0) if item.get('stat') else 0,
                    item.get('stat', {}).get('share', 0) if item.get('stat') else 0,
                ))
                inserted += 1

            conn.commit()

            file_hash = calculate_file_hash(file_path)
            update_sync_state(file_path, file_hash, len(data))

            return inserted, skipped
        finally:
            conn.close()

    except Exception as e:
        logger.error(f"导入文件失败 {file_path}: {e}")
        return 0, 0


def incremental_import(db_path: Optional[str] = None, max_workers: int = 4) -> dict:
    """增量导入：只导入已更改的文件"""
    if db_path is None:
        db_path = get_output_path(config.get('db_file', 'bilibili_history.db'))

    json_root = get_output_path('history_by_date')
    changed_files = get_changed_files(json_root)

    if not changed_files:
        return {
            "status": "success",
            "message": "没有需要同步的文件",
            "imported": 0,
            "skipped": 0,
            "files_processed": 0,
        }

    total_inserted = 0
    total_skipped = 0
    files_processed = 0
    start_time = time.time()

    with ThreadPoolExecutor(max_workers=max_workers) as executor:
        futures = {
            executor.submit(import_single_json_file, f, db_path, False): f
            for f in changed_files
        }
        for future in as_completed(futures):
            file = futures[future]
            try:
                inserted, skipped = future.result()
                total_inserted += inserted
                total_skipped += skipped
                files_processed += 1
                if inserted > 0:
                    logger.info(f"导入完成: {file} - {inserted} 条新记录")
            except Exception as e:
                logger.error(f"导入失败: {file} - {e}")

    elapsed = time.time() - start_time
    return {
        "status": "success",
        "message": f"增量同步完成",
        "imported": total_inserted,
        "skipped": total_skipped,
        "files_processed": files_processed,
        "total_changed_files": len(changed_files),
        "elapsed_seconds": round(elapsed, 2),
    }


def full_import(db_path: Optional[str] = None, max_workers: int = 4) -> dict:
    """全量导入：重新导入所有文件"""
    if db_path is None:
        db_path = get_output_path(config.get('db_file', 'bilibili_history.db'))

    json_root = get_output_path('history_by_date')
    all_files = []
    if os.path.exists(json_root):
        for year_dir in sorted(os.listdir(json_root)):
            year_path = os.path.join(json_root, year_dir)
            if not os.path.isdir(year_path):
                continue
            for month_dir in sorted(os.listdir(year_path)):
                month_path = os.path.join(year_path, month_dir)
                if not os.path.isdir(month_path):
                    continue
                for day_file in sorted(os.listdir(month_path)):
                    if day_file.endswith('.json'):
                        all_files.append(os.path.join(month_path, day_file))

    if not all_files:
        return {
            "status": "success",
            "message": "没有找到 JSON 文件",
            "imported": 0,
            "files_processed": 0,
        }

    total_inserted = 0
    total_skipped = 0
    files_processed = 0
    start_time = time.time()

    with ThreadPoolExecutor(max_workers=max_workers) as executor:
        futures = {
            executor.submit(import_single_json_file, f, db_path, True): f
            for f in all_files
        }
        for future in as_completed(futures):
            file = futures[future]
            try:
                inserted, skipped = future.result()
                total_inserted += inserted
                total_skipped += skipped
                files_processed += 1
            except Exception as e:
                logger.error(f"导入失败: {file} - {e}")

    elapsed = time.time() - start_time
    return {
        "status": "success",
        "message": f"全量同步完成",
        "imported": total_inserted,
        "skipped": total_skipped,
        "files_processed": files_processed,
        "elapsed_seconds": round(elapsed, 2),
    }
