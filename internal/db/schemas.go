package db

import (
	"fmt"
	"time"
)

// All schemas below match the Python backend EXACTLY.
// Run EnsureSchemas() on startup to create missing tables/columns.

// EnsureAllSchemas creates all tables if they don't exist.
func EnsureAllSchemas() error {
	dbs := []func() error{
		ensureHistorySchemas,
		ensureVideoLibrarySchemas,
		ensureInvalidVideosSchema,
		ensureCommentSchemas,
		ensureDynamicSchemas,
		ensureFavoriteSchemas,
		ensureWatchlaterSchema,
		ensureLikeSchema,
		ensureSchedulerSchemas,
		ensureInteractionSchemas,
		ensureSyncStateSchema,
		ensureHistorySimpleSchema,
		ensureCategorySchema,
	}
	for _, fn := range dbs {
		if err := fn(); err != nil {
			return err
		}
	}
	// Popular video schemas are year-specific, create current year
	year := time.Now().Year()
	return ensurePopularSchemas(year)
}

// ==================== bilibili_history.db ====================

func ensureHistorySchemas() error {
	db, err := OpenHistoryDB()
	if err != nil {
		return err
	}
	// Create deleted_history table (in the same DB)
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS deleted_history (
			id INTEGER PRIMARY KEY,
			bvid TEXT NOT NULL,
			view_at INTEGER NOT NULL,
			delete_time INTEGER NOT NULL,
			UNIQUE(bvid, view_at)
		)
	`)
	return err
}

// ==================== video_library.db ====================

func ensureVideoLibrarySchemas() error {
	dbPath := OutputDir() + "/database/video_library.db"
	db, err := Open(dbPath)
	if err != nil {
		return err
	}

	tables := []string{
		`CREATE TABLE IF NOT EXISTS video_base_info (
			id INTEGER PRIMARY KEY,
			bvid TEXT NOT NULL UNIQUE,
			aid INTEGER NOT NULL,
			videos INTEGER DEFAULT 1,
			tid INTEGER, tid_v2 INTEGER,
			tname TEXT, tname_v2 TEXT,
			copyright INTEGER, pic TEXT,
			title TEXT NOT NULL,
			pubdate INTEGER, ctime INTEGER,
			desc TEXT, desc_v2 TEXT,
			state INTEGER DEFAULT 0,
			duration INTEGER, mission_id INTEGER,
			dynamic TEXT, cid INTEGER,
			season_id INTEGER, premiere INTEGER,
			teenage_mode INTEGER DEFAULT 0,
			is_chargeable_season INTEGER DEFAULT 0,
			is_story INTEGER DEFAULT 0,
			is_upower_exclusive INTEGER DEFAULT 0,
			is_upower_play INTEGER DEFAULT 0,
			is_upower_preview INTEGER DEFAULT 0,
			is_upower_exclusive_with_qa INTEGER DEFAULT 0,
			no_cache INTEGER DEFAULT 0,
			is_season_display INTEGER DEFAULT 0,
			like_icon TEXT,
			need_jump_bv INTEGER DEFAULT 0,
			disable_show_up_info INTEGER DEFAULT 0,
			is_story_play INTEGER DEFAULT 0,
			owner_mid INTEGER, owner_name TEXT, owner_face TEXT,
			stat_view INTEGER DEFAULT 0, stat_danmaku INTEGER DEFAULT 0,
			stat_reply INTEGER DEFAULT 0, stat_favorite INTEGER DEFAULT 0,
			stat_coin INTEGER DEFAULT 0, stat_share INTEGER DEFAULT 0,
			stat_like INTEGER DEFAULT 0, stat_dislike INTEGER DEFAULT 0,
			stat_his_rank INTEGER DEFAULT 0, stat_now_rank INTEGER DEFAULT 0,
			stat_evaluation TEXT,
			dimension_width INTEGER, dimension_height INTEGER,
			dimension_rotate INTEGER DEFAULT 0,
			rights_bp INTEGER DEFAULT 0, rights_elec INTEGER DEFAULT 0,
			rights_download INTEGER DEFAULT 0, rights_movie INTEGER DEFAULT 0,
			rights_pay INTEGER DEFAULT 0, rights_hd5 INTEGER DEFAULT 0,
			rights_no_reprint INTEGER DEFAULT 0, rights_autoplay INTEGER DEFAULT 0,
			rights_ugc_pay INTEGER DEFAULT 0, rights_is_cooperation INTEGER DEFAULT 0,
			rights_ugc_pay_preview INTEGER DEFAULT 0, rights_no_background INTEGER DEFAULT 0,
			rights_clean_mode INTEGER DEFAULT 0, rights_is_stein_gate INTEGER DEFAULT 0,
			rights_is_360 INTEGER DEFAULT 0, rights_no_share INTEGER DEFAULT 0,
			rights_arc_pay INTEGER DEFAULT 0, rights_free_watch INTEGER DEFAULT 0,
			argue_msg TEXT, argue_type INTEGER DEFAULT 0, argue_link TEXT,
			fetch_time INTEGER NOT NULL,
			update_time INTEGER DEFAULT 0
		)`,
		`CREATE TABLE IF NOT EXISTS video_pages (
			id INTEGER PRIMARY KEY,
			bvid TEXT NOT NULL, cid INTEGER NOT NULL,
			page INTEGER NOT NULL, part TEXT, duration INTEGER,
			from_source TEXT, vid TEXT, weblink TEXT,
			dimension_width INTEGER, dimension_height INTEGER,
			dimension_rotate INTEGER DEFAULT 0, first_frame TEXT,
			ctime INTEGER DEFAULT 0,
			UNIQUE(bvid, cid)
		)`,
		`CREATE TABLE IF NOT EXISTS video_tags (
			id INTEGER PRIMARY KEY,
			bvid TEXT NOT NULL, tag_id INTEGER NOT NULL,
			tag_name TEXT NOT NULL, music_id TEXT,
			tag_type TEXT, jump_url TEXT, cover TEXT,
			content TEXT, short_content TEXT,
			type INTEGER, state INTEGER,
			UNIQUE(bvid, tag_id)
		)`,
		`CREATE TABLE IF NOT EXISTS uploader_info (
			mid INTEGER PRIMARY KEY,
			name TEXT NOT NULL, sex TEXT, face TEXT,
			face_nft INTEGER DEFAULT 0, face_nft_type INTEGER DEFAULT 0,
			sign TEXT, rank TEXT, level INTEGER DEFAULT 0,
			regtime INTEGER DEFAULT 0, spacesta INTEGER DEFAULT 0,
			birthday TEXT, place TEXT, description TEXT,
			article INTEGER DEFAULT 0, fans INTEGER DEFAULT 0,
			friend INTEGER DEFAULT 0, attention INTEGER DEFAULT 0,
			official_role INTEGER DEFAULT 0, official_title TEXT, official_desc TEXT,
			official_type INTEGER DEFAULT 0,
			vip_type INTEGER DEFAULT 0, vip_status INTEGER DEFAULT 0,
			vip_due_date INTEGER DEFAULT 0, vip_pay_type INTEGER DEFAULT 0,
			vip_theme_type INTEGER DEFAULT 0, vip_avatar_subscript INTEGER DEFAULT 0,
			vip_nickname_color TEXT, vip_role INTEGER DEFAULT 0,
			vip_avatar_subscript_url TEXT,
			pendant_pid INTEGER DEFAULT 0, pendant_name TEXT,
			pendant_image TEXT, pendant_expire INTEGER DEFAULT 0,
			nameplate_nid INTEGER DEFAULT 0, nameplate_name TEXT,
			nameplate_image TEXT, nameplate_image_small TEXT,
			nameplate_level TEXT, nameplate_condition TEXT,
			is_senior_member INTEGER DEFAULT 0,
			following INTEGER DEFAULT 0, archive_count INTEGER DEFAULT 0,
			article_count INTEGER DEFAULT 0, like_num INTEGER DEFAULT 0,
			fetch_time INTEGER NOT NULL, update_time INTEGER DEFAULT 0
		)`,
		`CREATE TABLE IF NOT EXISTS video_honors (
			id INTEGER PRIMARY KEY,
			bvid TEXT NOT NULL, aid INTEGER NOT NULL,
			type INTEGER NOT NULL, desc TEXT,
			weekly_recommend_num INTEGER DEFAULT 0,
			UNIQUE(bvid, type)
		)`,
		`CREATE TABLE IF NOT EXISTS video_subtitles (
			id INTEGER PRIMARY KEY,
			bvid TEXT NOT NULL,
			allow_submit INTEGER DEFAULT 0,
			subtitle_id INTEGER, lan TEXT, lan_doc TEXT,
			is_lock INTEGER, subtitle_url TEXT,
			type INTEGER, ai_type INTEGER, ai_status INTEGER,
			UNIQUE(bvid, subtitle_id)
		)`,
		`CREATE TABLE IF NOT EXISTS related_videos (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			bvid TEXT NOT NULL,
			related_bvid TEXT NOT NULL,
			related_title TEXT,
			related_pic TEXT,
			related_owner_name TEXT,
			UNIQUE(bvid, related_bvid)
		)`,
	}

	for _, sql := range tables {
		if _, err := db.Exec(sql); err != nil {
			return fmt.Errorf("create video table: %w", err)
		}
	}
	return nil
}

// ==================== bilibili_comments.db ====================

func ensureCommentSchemas() error {
	dbPath := OutputDir() + "/database/bilibili_comments.db"
	db, err := Open(dbPath)
	if err != nil {
		return err
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS comments (
			rpid TEXT PRIMARY KEY,
			uid TEXT NOT NULL,
			message TEXT NOT NULL,
			time INTEGER NOT NULL,
			rank INTEGER NOT NULL,
			rootid TEXT,
			parentid TEXT,
			oid TEXT NOT NULL,
			type INTEGER NOT NULL,
			fetch_time INTEGER NOT NULL
		)
	`)
	if err != nil {
		return err
	}

	db.Exec("CREATE INDEX IF NOT EXISTS idx_comments_uid ON comments(uid)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_comments_time ON comments(time)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_comments_fetch_time ON comments(fetch_time)")

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS comment_users (
			uid TEXT PRIMARY KEY,
			first_fetch_time INTEGER NOT NULL,
			last_fetch_time INTEGER NOT NULL
		)
	`)
	return err
}

// ==================== bilibili_dynamic.db ====================

func ensureDynamicSchemas() error {
	dbPath := OutputDir() + "/database/bilibili_dynamic.db"
	db, err := Open(dbPath)
	if err != nil {
		return err
	}

	tables := []string{
		`CREATE TABLE IF NOT EXISTS dynamic_core (
			host_mid TEXT NOT NULL, id_str TEXT NOT NULL,
			type TEXT, visible INTEGER, publish_ts INTEGER,
			comment_id_str TEXT, comment_type INTEGER,
			rid_str TEXT, txt TEXT, author_name TEXT,
			bvid TEXT, title TEXT, cover TEXT, desc TEXT,
			article_title TEXT, article_covers TEXT,
			opus_title TEXT, opus_summary_text TEXT,
			media_locals TEXT, media_count INTEGER,
			live_media_locals TEXT, live_media_count INTEGER,
			fetch_time INTEGER NOT NULL,
			PRIMARY KEY (host_mid, id_str)
		)`,
		`CREATE TABLE IF NOT EXISTS dynamic_author (
			host_mid TEXT NOT NULL, id_str TEXT NOT NULL,
			author_mid TEXT, author_name TEXT, face TEXT,
			PRIMARY KEY (host_mid, id_str)
		)`,
		`CREATE TABLE IF NOT EXISTS dynamic_stat (
			host_mid TEXT NOT NULL, id_str TEXT NOT NULL,
			like_count INTEGER, comment_count INTEGER,
			repost_count INTEGER, view_count INTEGER,
			PRIMARY KEY (host_mid, id_str)
		)`,
		`CREATE TABLE IF NOT EXISTS dynamic_topic (
			host_mid TEXT NOT NULL, id_str TEXT NOT NULL,
			topic_name TEXT, jump_url TEXT,
			PRIMARY KEY (host_mid, id_str)
		)`,
		`CREATE TABLE IF NOT EXISTS major_opus_pics (
			host_mid TEXT NOT NULL, id_str TEXT NOT NULL,
			idx INTEGER NOT NULL, url TEXT NOT NULL,
			PRIMARY KEY (host_mid, id_str, idx)
		)`,
		`CREATE TABLE IF NOT EXISTS major_archive_jump_urls (
			host_mid TEXT NOT NULL, id_str TEXT NOT NULL,
			idx INTEGER NOT NULL, url TEXT NOT NULL,
			PRIMARY KEY (host_mid, id_str, idx)
		)`,
	}

	for _, sql := range tables {
		if _, err := db.Exec(sql); err != nil {
			return fmt.Errorf("create dynamic table: %w", err)
		}
	}

	db.Exec("CREATE INDEX IF NOT EXISTS idx_dynamic_core_publish_ts ON dynamic_core(publish_ts)")
	return nil
}

// ==================== bilibili_favorites.db ====================

func ensureFavoriteSchemas() error {
	dbPath := OutputDir() + "/database/bilibili_favorites.db"
	db, err := Open(dbPath)
	if err != nil {
		return err
	}

	tables := []string{
		`CREATE TABLE IF NOT EXISTS favorites_folder (
			id INTEGER PRIMARY KEY,
			media_id INTEGER NOT NULL, fid INTEGER NOT NULL,
			mid INTEGER NOT NULL, title TEXT NOT NULL,
			cover TEXT, attr INTEGER, intro TEXT,
			ctime INTEGER, mtime INTEGER, state INTEGER,
			media_count INTEGER, fav_state INTEGER,
			like_state INTEGER, fetch_time INTEGER,
			UNIQUE(media_id)
		)`,
		`CREATE TABLE IF NOT EXISTS favorites_creator (
			mid INTEGER PRIMARY KEY,
			name TEXT NOT NULL, face TEXT,
			followed INTEGER, vip_type INTEGER,
			vip_status INTEGER, fetch_time INTEGER
		)`,
		`CREATE TABLE IF NOT EXISTS favorites_content (
			id INTEGER PRIMARY KEY,
			media_id INTEGER NOT NULL, content_id INTEGER NOT NULL,
			type INTEGER NOT NULL, title TEXT NOT NULL,
			cover TEXT, bvid TEXT, intro TEXT,
			page INTEGER, duration INTEGER, upper_mid INTEGER,
			attr INTEGER, ctime INTEGER, pubtime INTEGER,
			fav_time INTEGER, link TEXT, fetch_time INTEGER,
			creator_name TEXT, creator_face TEXT, bv_id TEXT,
			collect INTEGER, play INTEGER, danmaku INTEGER,
			play_switch INTEGER, reply INTEGER, view_text_1 TEXT,
			first_cid INTEGER, media_list_link TEXT,
			UNIQUE(media_id, content_id)
		)`,
	}

	for _, sql := range tables {
		if _, err := db.Exec(sql); err != nil {
			return fmt.Errorf("create favorite table: %w", err)
		}
	}

	db.Exec("CREATE INDEX IF NOT EXISTS idx_favorites_folder_mid ON favorites_folder(mid)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_favorites_folder_mtime ON favorites_folder(mtime)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_favorites_content_media_id ON favorites_content(media_id)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_favorites_content_upper_mid ON favorites_content(upper_mid)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_favorites_content_fav_time ON favorites_content(fav_time)")
	return nil
}

// ==================== bilibili_watchlater.db ====================

func ensureWatchlaterSchema() error {
	dbPath := OutputDir() + "/database/bilibili_watchlater.db"
	db, err := Open(dbPath)
	if err != nil {
		return err
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS watchlater_videos (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			bvid TEXT NOT NULL UNIQUE, aid INTEGER,
			title TEXT NOT NULL, pic TEXT, desc TEXT,
			duration INTEGER DEFAULT 0, tid INTEGER DEFAULT 0,
			tname TEXT, owner_name TEXT,
			owner_mid INTEGER DEFAULT 0, owner_face TEXT,
			add_at INTEGER DEFAULT 0, pubdate INTEGER DEFAULT 0,
			view INTEGER DEFAULT 0, danmaku INTEGER DEFAULT 0,
			link TEXT, fetch_time INTEGER NOT NULL
		)
	`)
	if err != nil {
		return err
	}

	db.Exec("CREATE INDEX IF NOT EXISTS idx_wl_bvid ON watchlater_videos(bvid)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_wl_add_at ON watchlater_videos(add_at)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_wl_owner ON watchlater_videos(owner_name)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_wl_tid ON watchlater_videos(tid)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_wl_fetch_time ON watchlater_videos(fetch_time)")
	return nil
}

// ==================== bilibili_likes.db ====================

func ensureLikeSchema() error {
	dbPath := OutputDir() + "/database/bilibili_likes.db"
	db, err := Open(dbPath)
	if err != nil {
		return err
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS liked_videos (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			bvid TEXT NOT NULL UNIQUE, aid INTEGER,
			title TEXT NOT NULL, pic TEXT, desc TEXT,
			duration INTEGER DEFAULT 0, tid INTEGER DEFAULT 0,
			tname TEXT, owner_name TEXT,
			owner_mid INTEGER DEFAULT 0, owner_face TEXT,
			pubdate INTEGER DEFAULT 0, view INTEGER DEFAULT 0,
			danmaku INTEGER DEFAULT 0, like_count INTEGER DEFAULT 0,
			link TEXT, fetch_time INTEGER NOT NULL,
			is_seen INTEGER DEFAULT 0
		)
	`)
	if err != nil {
		return err
	}

	db.Exec("CREATE INDEX IF NOT EXISTS idx_liked_bvid ON liked_videos(bvid)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_liked_pubdate ON liked_videos(pubdate)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_liked_owner ON liked_videos(owner_name)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_liked_tid ON liked_videos(tid)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_liked_fetch_time ON liked_videos(fetch_time)")
	return nil
}

// ==================== scheduler.db ====================

func ensureSchedulerSchemas() error {
	db, err := OpenSchedulerDB()
	if err != nil {
		return err
	}

	tables := []string{
		`CREATE TABLE IF NOT EXISTS main_tasks (
			task_id TEXT PRIMARY KEY,
			name TEXT NOT NULL, endpoint TEXT NOT NULL,
			method TEXT DEFAULT 'GET', params TEXT,
			schedule_type TEXT NOT NULL, schedule_time TEXT,
			schedule_delay INTEGER, interval_value INTEGER,
			interval_unit TEXT, enabled INTEGER DEFAULT 1,
			task_type TEXT DEFAULT 'main',
			created_at TIMESTAMP DEFAULT (datetime('now','localtime')),
			last_modified TIMESTAMP DEFAULT (datetime('now','localtime'))
		)`,
		`CREATE TABLE IF NOT EXISTS sub_tasks (
			task_id TEXT PRIMARY KEY,
			parent_id TEXT NOT NULL, name TEXT NOT NULL,
			sequence_number INTEGER NOT NULL,
			endpoint TEXT NOT NULL, method TEXT DEFAULT 'GET',
			params TEXT, schedule_type TEXT DEFAULT 'daily',
			enabled INTEGER DEFAULT 1,
			created_at TIMESTAMP DEFAULT (datetime('now','localtime')),
			last_modified TIMESTAMP DEFAULT (datetime('now','localtime')),
			FOREIGN KEY (parent_id) REFERENCES main_tasks(task_id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS task_status (
			task_id TEXT PRIMARY KEY,
			last_run_time TEXT, next_run_time TEXT,
			last_status TEXT, total_runs INTEGER DEFAULT 0,
			success_runs INTEGER DEFAULT 0, fail_runs INTEGER DEFAULT 0,
			avg_duration REAL DEFAULT 0, last_error TEXT,
			tags TEXT, success_rate REAL DEFAULT 0,
			FOREIGN KEY (task_id) REFERENCES main_tasks(task_id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS sub_task_status (
			task_id TEXT PRIMARY KEY,
			last_run_time TEXT, next_run_time TEXT,
			last_status TEXT, total_runs INTEGER DEFAULT 0,
			success_runs INTEGER DEFAULT 0, fail_runs INTEGER DEFAULT 0,
			avg_duration REAL DEFAULT 0, last_error TEXT,
			tags TEXT, success_rate REAL DEFAULT 0,
			FOREIGN KEY (task_id) REFERENCES sub_tasks(task_id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS task_executions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			task_id TEXT NOT NULL, start_time TEXT NOT NULL,
			end_time TEXT, duration REAL, status TEXT NOT NULL,
			error_message TEXT, output TEXT, triggered_by TEXT,
			next_run_time TEXT,
			created_at TIMESTAMP DEFAULT (datetime('now','localtime')),
			FOREIGN KEY (task_id) REFERENCES main_tasks(task_id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS sub_task_executions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			task_id TEXT NOT NULL, start_time TEXT NOT NULL,
			end_time TEXT, duration REAL, status TEXT NOT NULL,
			error_message TEXT, output TEXT, triggered_by TEXT,
			next_run_time TEXT,
			created_at TIMESTAMP DEFAULT (datetime('now','localtime')),
			FOREIGN KEY (task_id) REFERENCES sub_tasks(task_id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS task_dependencies (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			task_id TEXT NOT NULL, depends_on TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT (datetime('now','localtime')),
			UNIQUE(task_id, depends_on)
		)`,
	}

	for _, sql := range tables {
		if _, err := db.Exec(sql); err != nil {
			return fmt.Errorf("create scheduler table: %w", err)
		}
	}
	return nil
}

// ==================== bilibili_interactions.db ====================

func ensureInteractionSchemas() error {
	dbPath := OutputDir() + "/database/bilibili_interactions.db"
	db, err := Open(dbPath)
	if err != nil {
		return err
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS interaction_records (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			dedupe_key TEXT NOT NULL UNIQUE,
			source TEXT NOT NULL,
			oid INTEGER DEFAULT 0, aid INTEGER DEFAULT 0,
			bvid TEXT, title TEXT NOT NULL, cover TEXT,
			author_mid INTEGER DEFAULT 0,
			author_name TEXT, author_face TEXT,
			tname TEXT, duration INTEGER DEFAULT 0,
			pubtime INTEGER DEFAULT 0, ctime INTEGER DEFAULT 0,
			action_time INTEGER DEFAULT 0,
			action_time_source TEXT DEFAULT 'unknown',
			effective_time INTEGER DEFAULT 0,
			media_id INTEGER DEFAULT 0, media_title TEXT,
			raw_json TEXT, fetch_time INTEGER NOT NULL
		)
	`)
	if err != nil {
		return err
	}

	db.Exec("CREATE INDEX IF NOT EXISTS idx_interaction_source ON interaction_records(source)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_interaction_bvid ON interaction_records(bvid)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_interaction_author_mid ON interaction_records(author_mid)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_interaction_action_time ON interaction_records(action_time)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_interaction_effective_time ON interaction_records(effective_time)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_interaction_media_id ON interaction_records(media_id)")

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS interaction_sync_state (
			source TEXT PRIMARY KEY,
			last_sync_time INTEGER NOT NULL,
			status TEXT NOT NULL, message TEXT,
			total_records INTEGER DEFAULT 0,
			inserted_count INTEGER DEFAULT 0,
			updated_count INTEGER DEFAULT 0,
			details_json TEXT
		)
	`)
	return err
}

// ==================== sync_state.db ====================

func ensureSyncStateSchema() error {
	dbPath := OutputDir() + "/database/sync_state.db"
	db, err := Open(dbPath)
	if err != nil {
		return err
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS file_sync_state (
			file_path TEXT PRIMARY KEY,
			file_hash TEXT NOT NULL,
			record_count INTEGER DEFAULT 0,
			last_sync_time INTEGER NOT NULL,
			last_modified_time INTEGER NOT NULL
		)
	`)
	if err != nil {
		return err
	}

	db.Exec("CREATE INDEX IF NOT EXISTS idx_sync_hash ON file_sync_state(file_hash)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_sync_time ON file_sync_state(last_sync_time)")
	return nil
}

// ==================== bilibili_history_simple.db ====================

func ensureHistorySimpleSchema() error {
	dbPath := OutputDir() + "/database/bilibili_history_simple.db"
	db, err := Open(dbPath)
	if err != nil {
		return err
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS history_videos (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			bvid TEXT NOT NULL, aid INTEGER,
			title TEXT NOT NULL, pic TEXT, desc TEXT,
			duration INTEGER DEFAULT 0, tid INTEGER DEFAULT 0,
			tname TEXT, owner_name TEXT,
			owner_mid INTEGER DEFAULT 0, owner_face TEXT,
			view_at INTEGER DEFAULT 0, progress INTEGER DEFAULT 0,
			view INTEGER DEFAULT 0, danmaku INTEGER DEFAULT 0,
			link TEXT, fetch_time INTEGER NOT NULL,
			UNIQUE(bvid, view_at)
		)
	`)
	if err != nil {
		return err
	}

	db.Exec("CREATE INDEX IF NOT EXISTS idx_hist_bvid ON history_videos(bvid)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_hist_view_at ON history_videos(view_at)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_hist_owner ON history_videos(owner_name)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_hist_tid ON history_videos(tid)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_hist_fetch_time ON history_videos(fetch_time)")
	return nil
}

// ==================== popular_YYYY.db ====================

func ensurePopularSchemas(year int) error {
	dbPath := fmt.Sprintf("%s/database/popular_%d.db", OutputDir(), year)
	db, err := Open(dbPath)
	if err != nil {
		return err
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS popular_videos (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			aid TEXT, bvid TEXT, title TEXT,
			pubdate INTEGER, ctime INTEGER, desc TEXT,
			videos INTEGER, tid INTEGER, tname TEXT,
			copyright INTEGER, pic TEXT, duration INTEGER,
			owner_mid INTEGER, owner_name TEXT, owner_face TEXT,
			view_count INTEGER, danmaku_count INTEGER,
			reply_count INTEGER, favorite_count INTEGER,
			coin_count INTEGER, share_count INTEGER,
			like_count INTEGER, dynamic TEXT, cid TEXT,
			dimension_width INTEGER, dimension_height INTEGER,
			dimension_rotate INTEGER,
			short_link TEXT, first_frame TEXT, pub_location TEXT,
			cover43 TEXT, tidv2 INTEGER, tnamev2 TEXT,
			pid_v2 INTEGER, pid_name_v2 TEXT,
			season_type INTEGER, is_ogv INTEGER,
			rights_bp INTEGER, rights_elec INTEGER,
			rights_download INTEGER, rights_movie INTEGER,
			rights_pay INTEGER, rights_hd5 INTEGER,
			rights_no_reprint INTEGER, rights_autoplay INTEGER,
			rights_ugc_pay INTEGER, rights_is_cooperation INTEGER,
			rights_ugc_pay_preview INTEGER, rights_no_background INTEGER,
			rights_arc_pay INTEGER, rights_pay_free_watch INTEGER,
			stat_view INTEGER, stat_danmaku INTEGER,
			stat_reply INTEGER, stat_favorite INTEGER,
			stat_coin INTEGER, stat_share INTEGER,
			stat_now_rank INTEGER, stat_his_rank INTEGER,
			stat_like INTEGER, stat_dislike INTEGER,
			stat_vv INTEGER, stat_fav_g INTEGER, stat_like_g INTEGER,
			rcmd_reason_content TEXT, rcmd_reason_corner_mark INTEGER,
			ogv_info TEXT, ai_rcmd TEXT, fetch_time INTEGER,
			UNIQUE(aid, bvid, fetch_time)
		)
	`)
	if err != nil {
		return err
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS fetch_records (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			fetch_time INTEGER, total_fetched INTEGER,
			pages_fetched INTEGER, success INTEGER,
			failed_to_save INTEGER DEFAULT 0,
			duplicates_skipped INTEGER DEFAULT 0
		)
	`)
	if err != nil {
		return err
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS popular_video_tracking (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			aid TEXT, bvid TEXT, title TEXT,
			first_seen INTEGER, last_seen INTEGER,
			is_active INTEGER, total_duration INTEGER,
			highest_rank INTEGER, lowest_rank INTEGER,
			appearances INTEGER DEFAULT 1,
			UNIQUE(aid, bvid)
		)
	`)
	return err
}

// ==================== video_library.db: invalid_videos ====================

func ensureInvalidVideosSchema() error {
	dbPath := OutputDir() + "/database/video_library.db"
	db, err := Open(dbPath)
	if err != nil {
		return err
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS invalid_videos (
			id INTEGER PRIMARY KEY,
			bvid TEXT UNIQUE,
			error_type TEXT,
			error_code INTEGER,
			error_message TEXT,
			raw_response TEXT,
			first_check_time INTEGER,
			last_check_time INTEGER,
			check_count INTEGER DEFAULT 1
		)
	`)
	return err
}

// ==================== bilibili_history.db: video_categories ====================

func ensureCategorySchema() error {
	db, err := OpenHistoryDB()
	if err != nil {
		return err
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS video_categories (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			main_category TEXT NOT NULL,
			sub_category TEXT NOT NULL,
			alias TEXT NOT NULL,
			tid INTEGER NOT NULL,
			image TEXT
		)
	`)
	return err
}

// Helper to get current unix timestamp
func nowUnix() int64 {
	return time.Now().Unix()
}
