"""测试稍后再看删除API的请求格式"""
import sys
import yaml
import os
import requests

def load_real_config():
    """加载真实配置"""
    for path in ['config/config.yaml', '/app/config/config.yaml']:
        if os.path.exists(path):
            with open(path) as f:
                return yaml.safe_load(f)
    return {}

def test_api(config):
    sessdata = str(config.get('SESSDATA', ''))
    bili_jct = str(config.get('bili_jct', ''))
    dede = str(config.get('DedeUserID', ''))

    sys.stderr.write(f"SESSDATA: {'有效' if sessdata and not sessdata.startswith('Cookie') else '占位符'} (len={len(sessdata)})\n")
    sys.stderr.write(f"bili_jct: {'有效' if bili_jct and not bili_jct.startswith('你的') else '占位符'} (len={len(bili_jct)})\n")
    sys.stderr.write(f"DedeUserID: {'有效' if dede and not dede.startswith('你的') else '占位符'}\n")

    cookies = {'SESSDATA': sessdata, 'bili_jct': bili_jct, 'DedeUserID': dede}
    headers = {
        'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36',
        'Referer': 'https://www.bilibili.com/',
        'Origin': 'https://www.bilibili.com',
    }

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

if __name__ == '__main__':
    config = load_real_config()
    test_api(config)
