#!/usr/bin/env python3
"""
将本地存储的文件迁移到腾讯云 COS，并更新数据库中所有 URL。

用法:
  python3 scripts/migrate_to_cos.py --dry-run   # 仅预览，不实际操作
  python3 scripts/migrate_to_cos.py              # 执行迁移
"""
import os
import sys
import sqlite3
import argparse
import mimetypes
from pathlib import Path
from qcloud_cos import CosConfig, CosS3Client

PROJECT_ROOT = Path(__file__).resolve().parent.parent
DB_PATH = PROJECT_ROOT / "data" / "drama_generator.db"
STORAGE_PATH = PROJECT_ROOT / "data" / "storage"

LOCAL_URL_PREFIXES = [
    "http://localhost:5678/static/",
    "http://192.168.31.10:5678/static/",
    "http://192.168.31.10:7860/static/",
]

TABLES_TO_UPDATE = [
    ("characters",          "image_url",      "id"),
    ("scenes",              "image_url",      "id"),
    ("props",               "image_url",      "id"),
    ("storyboards",         "composed_image", "id"),
    ("storyboards",         "video_url",      "id"),
    ("image_generations",   "image_url",      "id"),
    ("video_generations",   "video_url",      "id"),
    ("assets",              "url",            "id"),
    ("assets",              "thumbnail_url",  "id"),
    ("character_libraries", "image_url",      "id"),
    ("video_merges",        "merged_url",     "id"),
]


def init_cos_client():
    secret_id = os.environ.get("COS_SECRET_ID", "")
    secret_key = os.environ.get("COS_SECRET_KEY", "")
    region = os.environ.get("COS_REGION", "ap-shanghai")
    bucket = os.environ.get("COS_BUCKET", "cinemaker-1300086205")
    cdn_url = os.environ.get("COS_CDN_URL", "")

    if not secret_id or not secret_key:
        env_file = PROJECT_ROOT / ".env"
        if env_file.exists():
            for line in env_file.read_text().splitlines():
                line = line.strip()
                if not line or line.startswith("#"):
                    continue
                if "=" in line:
                    k, v = line.split("=", 1)
                    k, v = k.strip(), v.strip()
                    if k == "COS_SECRET_ID":
                        secret_id = v
                    elif k == "COS_SECRET_KEY":
                        secret_key = v
                    elif k == "COS_REGION":
                        region = v
                    elif k == "COS_BUCKET":
                        bucket = v
                    elif k == "COS_CDN_URL":
                        cdn_url = v

    if not secret_id or not secret_key:
        print("错误: 未找到 COS_SECRET_ID / COS_SECRET_KEY，请设置环境变量或 .env 文件")
        sys.exit(1)

    cos_url = f"https://{bucket}.cos.{region}.myqcloud.com"
    base_url = cdn_url.rstrip("/") if cdn_url else cos_url

    config = CosConfig(
        Region=region,
        SecretId=secret_id,
        SecretKey=secret_key,
    )
    client = CosS3Client(config)
    return client, bucket, base_url


def url_to_local_path(url: str) -> Path | None:
    """从本地 URL 提取对应的文件路径"""
    for prefix in LOCAL_URL_PREFIXES:
        if url.startswith(prefix):
            relative = url[len(prefix):]
            local_path = STORAGE_PATH / relative
            if local_path.exists():
                return local_path
            break
    return None


def upload_file(client, bucket: str, local_path: Path, object_key: str, dry_run: bool) -> bool:
    if dry_run:
        return True
    content_type = mimetypes.guess_type(str(local_path))[0] or "application/octet-stream"
    try:
        client.upload_file(
            Bucket=bucket,
            LocalFilePath=str(local_path),
            Key=object_key,
            PartSize=10,
            MAXThread=5,
            ContentType=content_type,
        )
        return True
    except Exception as e:
        print(f"  上传失败: {object_key} -> {e}")
        return False


def migrate():
    parser = argparse.ArgumentParser(description="迁移本地存储到腾讯云 COS")
    parser.add_argument("--dry-run", action="store_true", help="仅预览，不实际操作")
    args = parser.parse_args()

    client, bucket, base_url = init_cos_client()
    print(f"COS Bucket: {bucket}")
    print(f"COS Base URL: {base_url}")
    print(f"Dry run: {args.dry_run}")
    print()

    # Phase 1: 扫描本地文件并上传到 COS
    print("=" * 60)
    print("Phase 1: 上传本地文件到 COS")
    print("=" * 60)

    uploaded_map = {}  # local_path -> cos_url
    all_files = list(STORAGE_PATH.rglob("*"))
    all_files = [f for f in all_files if f.is_file()]
    print(f"本地文件总数: {len(all_files)}")

    success_count = 0
    skip_count = 0
    fail_count = 0

    for i, local_path in enumerate(all_files, 1):
        relative = local_path.relative_to(STORAGE_PATH)
        object_key = str(relative)
        cos_url = f"{base_url}/{object_key}"

        if i % 50 == 0 or i == len(all_files):
            print(f"  进度: {i}/{len(all_files)}")

        if upload_file(client, bucket, local_path, object_key, args.dry_run):
            uploaded_map[str(local_path)] = cos_url
            success_count += 1
        else:
            fail_count += 1

    print(f"\n上传完成: 成功 {success_count}, 跳过 {skip_count}, 失败 {fail_count}")

    # Phase 2: 构建 URL 映射表 (old_url -> new_cos_url)
    url_map = {}
    for prefix in LOCAL_URL_PREFIXES:
        for local_path_str, cos_url in uploaded_map.items():
            local_path = Path(local_path_str)
            relative = local_path.relative_to(STORAGE_PATH)
            old_url = prefix + str(relative)
            url_map[old_url] = cos_url

    # Phase 3: 更新数据库 URL
    print()
    print("=" * 60)
    print("Phase 2: 更新数据库 URL")
    print("=" * 60)

    conn = sqlite3.connect(str(DB_PATH))
    cursor = conn.cursor()
    total_updated = 0

    for table, column, pk in TABLES_TO_UPDATE:
        try:
            cursor.execute(f"SELECT {pk}, {column} FROM {table} WHERE {column} IS NOT NULL AND {column} != ''")
        except Exception:
            continue

        rows = cursor.fetchall()
        update_count = 0

        for row_id, old_url in rows:
            if not old_url:
                continue

            new_url = url_map.get(old_url)
            if not new_url:
                local_path = url_to_local_path(old_url)
                if local_path and str(local_path) in uploaded_map:
                    new_url = uploaded_map[str(local_path)]

            if new_url and new_url != old_url:
                if not args.dry_run:
                    cursor.execute(f"UPDATE {table} SET {column} = ? WHERE {pk} = ?", (new_url, row_id))
                update_count += 1

        if update_count > 0:
            print(f"  {table}.{column}: {update_count} 条记录更新")
            total_updated += update_count

    if not args.dry_run:
        conn.commit()
    conn.close()

    print(f"\n数据库更新完成: 共 {total_updated} 条记录")

    if args.dry_run:
        print("\n[DRY RUN] 以上为预览，未实际执行任何操作。去掉 --dry-run 执行实际迁移。")
    else:
        print("\n迁移完成！请重启服务使 COS 配置生效: bash start.sh restart")


if __name__ == "__main__":
    migrate()
