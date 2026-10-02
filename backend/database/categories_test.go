package database

import "testing"

func TestEnsureCategoriesTable(t *testing.T) {
	if err := EnsureCategoriesTable(); err != nil {
		t.Fatal(err)
	}
	// 幂等：表已存在时直接返回
	if err := EnsureCategoriesTable(); err != nil {
		t.Fatal(err)
	}
	var n int
	if err := testConn().QueryRow("SELECT COUNT(*) FROM video_categories").Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n < 50 {
		t.Fatalf("InitCategories 应插入内置分类数据, got %d", n)
	}
}

func TestGetCategories(t *testing.T) {
	nodes, err := GetCategories()
	if err != nil {
		t.Fatal(err)
	}
	if len(nodes) == 0 {
		t.Fatal("至少应有一个主分类")
	}
	seen := map[string]bool{}
	subsTotal := 0
	for _, node := range nodes {
		if node.Name == "" {
			t.Fatal("主分类名不能为空")
		}
		if seen[node.Name] {
			t.Fatalf("主分类重复: %s", node.Name)
		}
		seen[node.Name] = true
		subSeen := map[string]bool{}
		for _, sub := range node.SubCategories {
			if sub.Name == node.Name {
				t.Fatalf("子分类不应与主分类同名: %s", sub.Name)
			}
			if subSeen[sub.Name] {
				t.Fatalf("子分类去重失效: %s/%s", node.Name, sub.Name)
			}
			subSeen[sub.Name] = true
			subsTotal++
		}
	}
	if subsTotal == 0 {
		t.Fatal("应存在子分类")
	}
}

func TestGetMainCategories(t *testing.T) {
	mains, err := GetMainCategories()
	if err != nil {
		t.Fatal(err)
	}
	if len(mains) == 0 {
		t.Fatal("主分类列表不能为空")
	}
	seen := map[string]bool{}
	for _, m := range mains {
		name := m["name"]
		if name == "" {
			t.Fatalf("缺少 name: %#v", m)
		}
		if seen[name] {
			t.Fatalf("主分类重复: %s", name)
		}
		seen[name] = true
	}
}

func TestGetSubCategories(t *testing.T) {
	mains, err := GetMainCategories()
	if err != nil {
		t.Fatal(err)
	}
	main := mains[0]["name"]
	subs, err := GetSubCategories(main)
	if err != nil {
		t.Fatal(err)
	}
	for _, s := range subs {
		if s.Name == main {
			t.Fatal("子分类不能与主分类同名")
		}
	}
	// 未知主分类应返回空而非报错
	subs, err = GetSubCategories("__不存在的分类__")
	if err != nil {
		t.Fatal(err)
	}
	if len(subs) != 0 {
		t.Fatalf("未知分类应返回空, got %d", len(subs))
	}
}
