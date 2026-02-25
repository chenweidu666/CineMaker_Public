#!/usr/bin/env python3
"""批量设置 COS 对象为 public-read"""
import os
import sys
from pathlib import Path
from qcloud_cos import CosConfig, CosS3Client

PROJECT_ROOT = Path(__file__).resolve().parent.parent
STORAGE_PATH = PROJECT_ROOT / "data" / "storage"


def load_env():
    env_file = PROJECT_ROOT / ".env"
    cfg = {}
    if env_file.exists():
        for line in env_file.read_text().splitlines():
            line = line.strip()
            if not line or line.startswith("#"):
                continue
            if "=" in line:
                k, v = line.split("=", 1)
                cfg[k.strip()] = v.strip()
    return cfg


def main():
    cfg = load_env()
    secret_id = os.environ.get("COS_SECRET_ID") or cfg.get("COS_SECRET_ID", "")
    secret_key = os.environ.get("COS_SECRET_KEY") or cfg.get("COS_SECRET_KEY", "")
    region = os.environ.get("COS_REGION") or cfg.get("COS_REGION", "ap-shanghai")
    bucket = os.environ.get("COS_BUCKET") or cfg.get("COS_BUCKET", "cinemaker-1300086205")

    if not secret_id or not secret_key:
        print("错误: 缺少 COS 密钥")
        sys.exit(1)

    cos_config = CosConfig(Region=region, SecretId=secret_id, SecretKey=secret_key)
    client = CosS3Client(cos_config)

    all_files = [f for f in STORAGE_PATH.rglob("*") if f.is_file()]
    print(f"需要设置 ACL 的文件: {len(all_files)}")

    success = 0
    fail = 0
    for i, local_path in enumerate(all_files, 1):
        relative = local_path.relative_to(STORAGE_PATH)
        object_key = str(relative)
        try:
            client.put_object_acl(Bucket=bucket, Key=object_key, ACL="public-read")
            success += 1
        except Exception as e:
            print(f"  失败: {object_key} -> {e}")
            fail += 1

        if i % 50 == 0 or i == len(all_files):
            print(f"  进度: {i}/{len(all_files)} (成功: {success}, 失败: {fail})")

    print(f"\n完成: 成功 {success}, 失败 {fail}")


if __name__ == "__main__":
    main()
