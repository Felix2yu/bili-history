"""测试稍后再看删除API的请求格式"""
import sys
import yaml
import os
import requests
import importlib.util

def load_real_config():
    """加载真实配置"""
    for path in ['config/config.yaml', '/app/config/config.yaml']:
        if os.path.exists(path):
            with open(path) as f:
                return yaml.safe_load(f)
    return {}

def load_wbi_sign():
    """动态加载 wbi_sign 模块"""
    for path in ['/app/scripts/wbi_sign.py', 'scripts/wbi_sign.py']:
        if os.path.exists(path):
            spec = importlib.util.spec_from_file_location("wbi_sign", path)
            mod = importlib.util.module_from_spec(spec)
            spec.loader.exec_module(mod)
            return mod
    return None

def test_api(config):
    sessdata = str(config.get('SESSDATA', ''))
    bili_jct = str(config.get('bili_jct', ''))
    dede = str(config.get('DedeUserID', ''))

    sys.stderr.write(f"SESSDATA: len={len(sessdata)}, hasPercent={'%' in sessdata}, sample={sessdata[:8]}...{sessdata[-8:]}\n")
    sys.stderr.write(f"bili_jct: len={len(bili_jct)}, sample={bili_jct[:4]}...{bili_jct[-4:]}\n")
    sys.stderr.write(f"DedeUserID: {dede}\n")

    cookies = {'SESSDATA': sessdata, 'bili_jct': bili_jct, 'DedeUserID': dede}
    headers = {
        'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36',
        'Referer': 'https://www.bilibili.com/',
        'Origin': 'https://www.bilibili.com',
    }

    # 测试: nav API 认证状态
    r_nav_test = requests.get('https://api.bilibili.com/x/web-interface/nav',
        cookies=cookies, headers=headers)
    j_nav_test = r_nav_test.json()
    sys.stderr.write(f"nav: code={j_nav_test.get('code')} isLogin={j_nav_test.get('data',{}).get('isLogin')} uid={j_nav_test.get('data',{}).get('uid')}\n\n")

    # 1. 获取列表确认凭证有效
    r0 = requests.get('https://api.bilibili.com/x/v2/history/toview', cookies=cookies, headers=headers)
    j0 = r0.json()
    sys.stderr.write(f"\n[1] 获取列表: code={j0.get('code')} msg={j0.get('message')}\n")

    if j0.get('code') != 0:
        sys.stderr.write("凭证无效，无法测试\n")
        return

    wl = j0.get('data', {}).get('list', [])
    sys.stderr.write(f"列表数量: {len(wl)}\n")
    if not wl:
        sys.stderr.write("列表为空，无法测试删除\n")
        return

    test_bvid = wl[-1]['bvid']
    sys.stderr.write(f"测试删除 bvid: {test_bvid}\n\n")

    # 额外验证: 检查 nav API 是否返回 wbi_img (确认会话完整)
    r_nav = requests.get('https://api.bilibili.com/x/web-interface/nav', cookies=cookies, headers=headers)
    j_nav = r_nav.json()
    wbi = j_nav.get('data', {}).get('wbi_img', {})
    sys.stderr.write(f"nav wbi_img: img={bool(wbi.get('img_url'))} sub={bool(wbi.get('sub_url'))}\n")
    sys.stderr.write(f"nav uid: {j_nav.get('data', {}).get('uid')}\n\n")

    # 方式1: form data + cookies参数
    r1 = requests.post('https://api.bilibili.com/x/v2/history/toview/del',
        data={'bvid': test_bvid, 'csrf': bili_jct},
        cookies=cookies, headers=headers)
    sys.stderr.write(f"[方式1] form+cookies: code={r1.json().get('code')} msg={r1.json().get('message')}\n")

    # 如果方式1失败，尝试其他方式
    if r1.json().get('code') != 0:
        # 方式2: JSON body
        headers2 = {**headers, 'Content-Type': 'application/json'}
        r2 = requests.post('https://api.bilibili.com/x/v2/history/toview/del',
            json={'bvid': test_bvid, 'csrf': bili_jct},
            cookies=cookies, headers=headers2)
        sys.stderr.write(f"[方式2] json+cookies: code={r2.json().get('code')} msg={r2.json().get('message')}\n")

        # 方式3: Cookie header手动拼接
        cookie_str = f"SESSDATA={sessdata}; bili_jct={bili_jct}; DedeUserID={dede}"
        headers3 = {**headers, 'Cookie': cookie_str}
        r3 = requests.post('https://api.bilibili.com/x/v2/history/toview/del',
            data={'bvid': test_bvid, 'csrf': bili_jct},
            headers=headers3)
        sys.stderr.write(f"[方式3] 手动Cookie: code={r3.json().get('code')} msg={r3.json().get('message')}\n")

        # 方式4: 不带 Origin
        headers4 = {k: v for k, v in headers.items() if k != 'Origin'}
        r4 = requests.post('https://api.bilibili.com/x/v2/history/toview/del',
            data={'bvid': test_bvid, 'csrf': bili_jct},
            cookies=cookies, headers=headers4)
        sys.stderr.write(f"[方式4] 无Origin: code={r4.json().get('code')} msg={r4.json().get('message')}\n")

        # 方式5: bvid 放 query, csrf 放 body
        r5 = requests.post(f'https://api.bilibili.com/x/v2/history/toview/del?bvid={test_bvid}',
            data={'csrf': bili_jct},
            cookies=cookies, headers=headers)
        sys.stderr.write(f"[方式5] bvid在query: code={r5.json().get('code')} msg={r5.json().get('message')}\n")

        # 方式6: 全部放 query 参数
        r6 = requests.post(f'https://api.bilibili.com/x/v2/history/toview/del?bvid={test_bvid}&csrf={bili_jct}',
            cookies=cookies, headers=headers)
        sys.stderr.write(f"[方式6] 全query: code={r6.json().get('code')} msg={r6.json().get('message')}\n")

        # 方式7: JSON + Session
        s = requests.Session()
        s.cookies.update(cookies)
        r7 = s.post('https://api.bilibili.com/x/v2/history/toview/del',
            json={'bvid': test_bvid, 'csrf': bili_jct},
            headers={k: v for k, v in headers.items()} | {'Content-Type': 'application/json'})
        sys.stderr.write(f"[方式7] json+session: code={r7.json().get('code')} msg={r7.json().get('message')}\n")

        # 方式8: JSON body，csrf 只放 cookie 不放 body
        r8 = requests.post('https://api.bilibili.com/x/v2/history/toview/del',
            json={'bvid': test_bvid},
            cookies=cookies, headers={**headers, 'Content-Type': 'application/json'})
        sys.stderr.write(f"[方式8] json无csrf-body: code={r8.json().get('code')} msg={r8.json().get('message')}\n")

        # 方式9: JSON + WBI 签名
        wbi_mod = load_wbi_sign()
        if wbi_mod:
            try:
                params = {'bvid': test_bvid, 'csrf': bili_jct}
                signed = wbi_mod.get_wbi_sign(params)
                r9 = requests.post('https://api.bilibili.com/x/v2/history/toview/del',
                    json=signed,
                    cookies=cookies, headers={**headers, 'Content-Type': 'application/json'})
                sys.stderr.write(f"[方式9] json+wbi: code={r9.json().get('code')} msg={r9.json().get('message')}\n")
            except Exception as e:
                sys.stderr.write(f"[方式9] json+wbi: error={e}\n")
        else:
            sys.stderr.write("[方式9] json+wbi: wbi_sign模块加载失败\n")

        # 方式10: form + WBI 签名
        if wbi_mod:
            try:
                params2 = {'bvid': test_bvid, 'csrf': bili_jct}
                signed2 = wbi_mod.get_wbi_sign(params2)
                r10 = requests.post('https://api.bilibili.com/x/v2/history/toview/del',
                    data=signed2,
                    cookies=cookies, headers=headers)
                sys.stderr.write(f"[方式10] form+wbi: code={r10.json().get('code')} msg={r10.json().get('message')}\n")
            except Exception as e:
                sys.stderr.write(f"[方式10] form+wbi: error={e}\n")
        else:
            sys.stderr.write("[方式10] form+wbi: wbi_sign模块加载失败\n")

if __name__ == '__main__':
    config = load_real_config()
    test_api(config)
