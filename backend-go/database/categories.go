package database

import (
	"database/sql"
	"fmt"

	"bilibili-history-go/models"
	"bilibili-history-go/utils"
)

func EnsureCategoriesTable() error {
	db := GetSQLiteDB()
	conn := db.GetDB()
	if conn == nil {
		return fmt.Errorf("database not initialized")
	}

	exists, _ := db.TableExists("video_categories")
	if exists {
		return nil
	}

	return InitCategories()
}

func InitCategories() error {
	db := GetSQLiteDB()
	conn := db.GetDB()
	if conn == nil {
		return fmt.Errorf("database not initialized")
	}

	createSQL := `
	CREATE TABLE IF NOT EXISTS video_categories (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		main_category TEXT NOT NULL,
		sub_category TEXT NOT NULL,
		alias TEXT,
		tid INTEGER,
		image TEXT
	);
	`

	_, err := conn.Exec(createSQL)
	if err != nil {
		return fmt.Errorf("failed to create video_categories table: %v", err)
	}

	categories := []struct {
		mainCategory string
		subCategory  string
		alias        string
		tid          int
		image        string
	}{
		{"动画", "MAD·AMV", "MAD", 24, ""},
		{"动画", "MMD·3D", "MMD", 25, ""},
		{"动画", "短片·手书·配音", "短片", 47, ""},
		{"动画", "手办·模玩", "手办模玩", 210, ""},
		{"动画", "特摄", "特摄", 86, ""},
		{"动画", "动漫杂谈", "动漫杂谈", 253, ""},
		{"动画", "综合", "动画综合", 27, ""},
		{"番剧", "连载动画", "连载动画", 33, ""},
		{"番剧", "完结动画", "完结动画", 32, ""},
		{"番剧", "资讯", "番剧资讯", 51, ""},
		{"番剧", "官方延伸", "官方延伸", 152, ""},
		{"番剧", "新番时间表", "新番时间表", 169, ""},
		{"番剧", "番剧索引", "番剧索引", 170, ""},
		{"国创", "国产动画", "国产动画", 153, ""},
		{"国创", "国产原创相关", "国产原创", 168, ""},
		{"国创", "布袋戏", "布袋戏", 169, ""},
		{"国创", "动态漫·广播剧", "动态漫", 195, ""},
		{"国创", "资讯", "国创资讯", 170, ""},
		{"国创", "全部分区", "国创全部分区", 171, ""},
		{"音乐", "原创音乐", "原创音乐", 28, ""},
		{"音乐", "翻唱", "翻唱", 31, ""},
		{"音乐", "VOCALOID·UTAU", "V家", 30, ""},
		{"音乐", "电音", "电音", 194, ""},
		{"音乐", "演奏", "演奏", 59, ""},
		{"音乐", "MV", "MV", 193, ""},
		{"音乐", "音乐现场", "现场", 29, ""},
		{"音乐", "音乐综合", "音乐综合", 130, ""},
		{"音乐", "说唱", "说唱", 127, ""},
		{"音乐", "乐评盘点", "乐评盘点", 243, ""},
		{"音乐", "音乐教学", "音乐教学", 244, ""},
		{"舞蹈", "宅舞", "宅舞", 20, ""},
		{"舞蹈", "街舞", "街舞", 198, ""},
		{"舞蹈", "明星舞蹈", "明星舞蹈", 199, ""},
		{"舞蹈", "中国舞", "中国舞", 200, ""},
		{"舞蹈", "舞蹈综合", "舞蹈综合", 154, ""},
		{"舞蹈", "舞蹈教程", "舞蹈教程", 156, ""},
		{"游戏", "单机游戏", "单机游戏", 17, ""},
		{"游戏", "网络游戏", "网络游戏", 65, ""},
		{"游戏", "手机游戏", "手机游戏", 171, ""},
		{"游戏", "电子竞技", "电子竞技", 172, ""},
		{"游戏", "桌游棋牌", "桌游棋牌", 173, ""},
		{"游戏", "GMV", "GMV", 121, ""},
		{"游戏", "音游", "音游", 136, ""},
		{"游戏", "MUGEN", "MUGEN", 19, ""},
		{"游戏", "游戏赛事", "游戏赛事", 74, ""},
		{"游戏", "游戏综合", "游戏综合", 174, ""},
		{"知识", "科学科普", "科普", 201, ""},
		{"知识", "社科·法律·心理", "社科心理", 124, ""},
		{"知识", "人文历史", "人文历史", 228, ""},
		{"知识", "商业财经", "财经", 207, ""},
		{"知识", "校园学习", "校园学习", 208, ""},
		{"知识", "职业职场", "职场", 209, ""},
		{"知识", "设计·创意", "设计创意", 229, ""},
		{"知识", "技能", "技能", 122, ""},
		{"科技", "数码", "数码", 95, ""},
		{"科技", "软件应用", "软件", 230, ""},
		{"科技", "计算机技术", "计算机", 231, ""},
		{"科技", "极客DIY", "极客DIY", 232, ""},
		{"科技", "工业·制造·机械", "工业制造", 233, ""},
		{"科技", "汽车", "汽车", 234, ""},
		{"运动", "健身", "健身", 164, ""},
		{"运动", "运动文化", "运动文化", 235, ""},
		{"运动", "篮球·足球", "篮球足球", 236, ""},
		{"运动", "极限运动·街头运动", "极限运动", 237, ""},
		{"运动", "竞技体育", "竞技体育", 238, ""},
		{"运动", "运动综合", "运动综合", 239, ""},
		{"汽车", "汽车生活", "汽车生活", 240, ""},
		{"汽车", "汽车科普", "汽车科普", 241, ""},
		{"汽车", "汽车技术", "汽车技术", 242, ""},
		{"汽车", "改装", "改装", 245, ""},
		{"汽车", "新车资讯", "新车资讯", 246, ""},
		{"汽车", "二手车", "二手车", 247, ""},
		{"汽车", "汽车评测", "汽车评测", 248, ""},
		{"生活", "搞笑", "搞笑", 138, ""},
		{"生活", "日常", "日常", 21, ""},
		{"生活", "美食圈", "美食", 211, ""},
		{"生活", "手工", "手工", 161, ""},
		{"生活", "绘画", "绘画", 162, ""},
		{"生活", "运动", "生活运动", 163, ""},
		{"生活", "汽车", "生活汽车", 164, ""},
		{"生活", "宠物", "宠物", 175, ""},
		{"生活", "三农", "三农", 176, ""},
		{"生活", "家居房产", "家居房产", 177, ""},
		{"生活", "手工", "生活手工", 178, ""},
		{"生活", "绘画", "生活绘画", 179, ""},
		{"生活", "日常", "生活日常", 21, ""},
		{"美食制作", "美食制作", "美食制作", 211, ""},
		{"美食侦探", "美食侦探", "美食侦探", 212, ""},
		{"美食测评", "美食测评", "美食测评", 213, ""},
		{"田园美食", "田园美食", "田园美食", 214, ""},
		{"美食记录", "美食记录", "美食记录", 215, ""},
		{"动物圈", "喵星人", "猫", 217, ""},
		{"动物圈", "汪星人", "狗", 218, ""},
		{"动物圈", "大熊猫", "熊猫", 219, ""},
		{"动物圈", "野生动物", "野生动物", 220, ""},
		{"动物圈", "爬宠", "爬宠", 221, ""},
		{"动物圈", "动物综合", "动物综合", 222, ""},
		{"鬼畜", "鬼畜调教", "鬼畜调教", 22, ""},
		{"鬼畜", "音MAD", "音MAD", 26, ""},
		{"鬼畜", "人力VOCALOID", "人力V", 126, ""},
		{"鬼畜", "鬼畜剧场", "鬼畜剧场", 216, ""},
		{"鬼畜", "教程演示", "鬼畜教程", 127, ""},
		{"时尚", "美妆护肤", "美妆", 157, ""},
		{"时尚", "仿妆cos", "仿妆", 250, ""},
		{"时尚", "穿搭", "穿搭", 158, ""},
		{"时尚", "时尚潮流", "时尚潮流", 159, ""},
		{"时尚", "发型", "发型", 251, ""},
		{"资讯", "热点", "热点", 252, ""},
		{"资讯", "科技", "资讯科技", 124, ""},
		{"资讯", "社会", "社会资讯", 125, ""},
		{"资讯", "娱乐", "娱乐资讯", 126, ""},
		{"资讯", "专栏", "专栏", 127, ""},
		{"娱乐", "综艺", "综艺", 71, ""},
		{"娱乐", "明星", "明星", 137, ""},
		{"娱乐", "娱乐杂谈", "娱乐杂谈", 249, ""},
		{"娱乐", "粉丝创作", "粉丝创作", 250, ""},
		{"娱乐", "娱乐综合", "娱乐综合", 251, ""},
		{"影视", "影视杂谈", "影视杂谈", 182, ""},
		{"影视", "影视剪辑", "影视剪辑", 183, ""},
		{"影视", "小剧场", "小剧场", 184, ""},
		{"影视", "短片", "影视短片", 85, ""},
		{"影视", "预告·资讯", "预告资讯", 185, ""},
		{"影视", "影视综合", "影视综合", 186, ""},
		{"纪录片", "人文·历史", "人文历史纪录片", 177, ""},
		{"纪录片", "科学·探索·自然", "自然纪录片", 178, ""},
		{"纪录片", "军事", "军事纪录片", 179, ""},
		{"纪录片", "社会·美食·旅行", "社会纪录片", 180, ""},
		{"电影", "华语电影", "华语电影", 147, ""},
		{"电影", "欧美电影", "欧美电影", 148, ""},
		{"电影", "日本电影", "日本电影", 149, ""},
		{"电影", "韩国电影", "韩国电影", 183, ""},
		{"电影", "其他国家", "其他国家电影", 184, ""},
		{"电视剧", "国产剧", "国产剧", 185, ""},
		{"电视剧", "海外剧", "海外剧", 187, ""},
		{"VLOG", "VLOG", "VLOG", 250, ""},
		{"生活记录", "生活记录", "生活记录", 251, ""},
		{"搞笑视频", "搞笑视频", "搞笑视频", 252, ""},
		{"搞笑日常", "搞笑日常", "搞笑日常", 253, ""},
		{"搞笑脑洞", "搞笑脑洞", "搞笑脑洞", 254, ""},
		{"动物圈", "动物圈", "动物圈", 217, ""},
	}

	insertSQL := `INSERT OR IGNORE INTO video_categories (main_category, sub_category, alias, tid, image) VALUES (?, ?, ?, ?, ?)`

	tx, err := conn.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare(insertSQL)
	if err != nil {
		tx.Rollback()
		return err
	}
	defer stmt.Close()

	for _, cat := range categories {
		_, err = stmt.Exec(cat.mainCategory, cat.subCategory, cat.alias, cat.tid, cat.image)
		if err != nil {
			utils.LogWarning("Failed to insert category: %v", err)
		}
	}

	return tx.Commit()
}

func GetCategories() ([]models.CategoryNode, error) {
	if err := EnsureCategoriesTable(); err != nil {
		return nil, err
	}

	db := GetSQLiteDB()
	conn := db.GetDB()
	if conn == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	rows, err := conn.Query(`
		SELECT main_category, sub_category, alias, tid, image
		FROM video_categories
		ORDER BY main_category, sub_category
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	categoryMap := make(map[string]*models.CategoryNode)

	for rows.Next() {
		var mainCat, subCat, alias string
		var tid int
		var image sql.NullString

		if err := rows.Scan(&mainCat, &subCat, &alias, &tid, &image); err != nil {
			continue
		}

		if _, ok := categoryMap[mainCat]; !ok {
			categoryMap[mainCat] = &models.CategoryNode{
				Name:         mainCat,
				Image:        image.String,
				SubCategories: []models.SubCategory{},
			}
		}

		if subCat != mainCat {
			categoryMap[mainCat].SubCategories = append(categoryMap[mainCat].SubCategories, models.SubCategory{
				Name:  subCat,
				Alias: alias,
				Tid:   tid,
			})
		}
	}

	var result []models.CategoryNode
	for _, cat := range categoryMap {
		result = append(result, *cat)
	}

	return result, nil
}

func GetMainCategories() ([]map[string]string, error) {
	if err := EnsureCategoriesTable(); err != nil {
		return nil, err
	}

	db := GetSQLiteDB()
	conn := db.GetDB()
	if conn == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	rows, err := conn.Query(`
		SELECT DISTINCT main_category, image
		FROM video_categories
		ORDER BY main_category
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []map[string]string
	for rows.Next() {
		var name string
		var image sql.NullString
		if err := rows.Scan(&name, &image); err != nil {
			continue
		}
		categories = append(categories, map[string]string{
			"name":  name,
			"image": image.String,
		})
	}

	return categories, nil
}

func GetSubCategories(mainCategory string) ([]models.SubCategory, error) {
	if err := EnsureCategoriesTable(); err != nil {
		return nil, err
	}

	db := GetSQLiteDB()
	conn := db.GetDB()
	if conn == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	rows, err := conn.Query(`
		SELECT sub_category, alias, tid
		FROM video_categories
		WHERE main_category = ? AND sub_category != main_category
		ORDER BY sub_category
	`, mainCategory)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []models.SubCategory
	for rows.Next() {
		var name, alias string
		var tid int
		if err := rows.Scan(&name, &alias, &tid); err != nil {
			continue
		}
		categories = append(categories, models.SubCategory{
			Name:  name,
			Alias: alias,
			Tid:   tid,
		})
	}

	return categories, nil
}
