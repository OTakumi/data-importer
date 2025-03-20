#!/usr/bin/env python3
"""
MongoDB JSONインポーター用のテストデータを生成するスクリプト

使い方:
    python generate_test_data.py

生成されるファイル:
    1. users_array.json - 配列形式の有効なJSONファイル
    2. product_object.json - 単一オブジェクト形式の有効なJSONファイル
    3. invalid.json - 不正なJSONファイル
    4. large_data.json - 大規模テスト用のJSONファイル
"""

import json
import os
import random
import string


# 保存先ディレクトリの作成
def ensure_dir():
    os.makedirs(os.path.dirname(os.path.abspath(__file__)), exist_ok=True)


# ランダムなユーザーデータを生成する関数
def generate_user(user_id):
    return {
        "id": user_id,
        "name": f"{random.choice(['田中', '鈴木', '佐藤', '伊藤', '渡辺'])} {random.choice(['太郎', '花子', '次郎', '幸子', '雄太'])}",
        "email": f"user{user_id}@example.com",
        "age": random.randint(18, 80),
        "active": random.choice([True, False]),
        "created_at": "2023-01-01T12:00:00Z",
        "tags": random.sample(["premium", "student", "business", "trial", "admin"], random.randint(1, 3)),
    }


# ランダムな製品データを生成する関数
def generate_product():
    return {
        "product_id": "PRD-" + "".join(random.choices(string.ascii_uppercase + string.digits, k=6)),
        "name": f"{random.choice(['高級', 'スタンダード', 'エコノミー'])} {random.choice(['椅子', 'テーブル', 'ソファ', 'ベッド', '本棚'])}",
        "price": round(random.uniform(1000, 50000), -1),  # 10円単位に丸める
        "in_stock": random.randint(0, 100),
        "categories": random.sample(["家具", "インテリア", "オフィス", "寝具", "収納"], random.randint(1, 3)),
        "details": {
            "color": random.choice(["ナチュラル", "ブラウン", "ホワイト", "ブラック", "グレー"]),
            "material": random.choice(["木材", "金属", "プラスチック", "ガラス", "布地"]),
            "dimensions": {
                "width": random.randint(30, 200),
                "height": random.randint(30, 200),
                "depth": random.randint(30, 200),
            },
        },
        "rating": round(random.uniform(1, 5), 1),
        "last_updated": "2023-03-15T09:30:00Z",
    }


def main():
    ensure_dir()
    script_dir = os.path.dirname(os.path.abspath(__file__))

    # 1. 配列形式のJSONファイル (users_array.json)
    users = [generate_user(i) for i in range(1, 11)]
    with open(os.path.join(script_dir, "users_array.json"), "w", encoding="utf-8") as f:
        json.dump(users, f, ensure_ascii=False, indent=2)
    print("users_array.json を生成しました (10ユーザー)")

    # 2. 単一オブジェクト形式のJSONファイル (product_object.json)
    product = generate_product()
    with open(os.path.join(script_dir, "product_object.json"), "w", encoding="utf-8") as f:
        json.dump(product, f, ensure_ascii=False, indent=2)
    print("product_object.json を生成しました (単一製品)")

    # 3. 不正なJSONファイル (invalid.json)
    with open(os.path.join(script_dir, "invalid.json"), "w", encoding="utf-8") as f:
        f.write('{"name": "不正なJSON", "description": "閉じ括弧がない')
    print("invalid.json を生成しました (不正なJSON)")

    # 4. 大規模テスト用のJSONファイル (large_data.json)
    large_users = [generate_user(i) for i in range(1, 10001)]  # 1万件のユーザーデータ
    with open(os.path.join(script_dir, "large_data.json"), "w", encoding="utf-8") as f:
        json.dump(large_users, f, ensure_ascii=False)  # インデントなしで保存
    print("large_data.json を生成しました (10,000ユーザー)")

    # 5. ネストされたデータのJSONファイル (nested_data.json)
    nested_data = {
        "company": "サンプル株式会社",
        "established": 1995,
        "locations": [
            {"name": "東京本社", "employees": 120},
            {"name": "大阪支店", "employees": 45},
            {"name": "名古屋支店", "employees": 30},
        ],
        "departments": {
            "営業部": {"manager": "山田太郎", "budget": 5000000},
            "開発部": {"manager": "佐藤花子", "budget": 8000000},
            "人事部": {"manager": "鈴木次郎", "budget": 3000000},
        },
        "products": [
            {"id": "P001", "name": "製品A", "price": 2000},
            {"id": "P002", "name": "製品B", "price": 3500},
            {"id": "P003", "name": "製品C", "price": 1800},
        ],
    }
    with open(os.path.join(script_dir, "nested_data.json"), "w", encoding="utf-8") as f:
        json.dump(nested_data, f, ensure_ascii=False, indent=2)
    print("nested_data.json を生成しました (ネスト構造)")


if __name__ == "__main__":
    main()
