#!/usr/bin/env python3
"""
CineMaker 剧本自动导入工具

解析标准格式 Markdown 剧本，自动创建数据库记录：
  - 章节 (Episode)
  - 角色造型 (Character + Outfit)
  - 场景 (Scene)
  - 道具 (Prop)
  - 分镜 (Storyboard) + 关联关系 (storyboard_characters, storyboard_props)

用法:
  python3 scripts/import_script.py docs/ai-vlogs/都市生活/EP03-xxx.md --drama-id 13
  python3 scripts/import_script.py EP03.md --drama-id 13 --dry-run
  python3 scripts/import_script.py EP03.md --drama-id 13 --parse-only
"""

import argparse
import os
import re
import sqlite3
import sys
from datetime import datetime
from dataclasses import dataclass, field
from typing import Dict, List, Optional


# ──────────────────────────── Data Classes ────────────────────────────

@dataclass
class CharacterEntry:
    name: str
    outfit: str
    shots: str
    brief: str
    full_name: str = ""
    appearance: str = ""
    is_lead: bool = False


@dataclass
class ShotSection:
    number: int
    title: str
    duration: int
    location: str = ""
    main_outfit: str = ""
    shot_type: str = ""
    characters: List[str] = field(default_factory=list)
    prev_shot_ref: str = ""
    first_frame: str = ""
    middle_action: str = ""
    last_frame: str = ""
    props: List[str] = field(default_factory=list)


@dataclass
class ParsedScript:
    title: str = ""
    ep_type: str = ""
    duration: int = 90
    style: str = ""
    extra_info: str = ""
    characters: List[CharacterEntry] = field(default_factory=list)
    shots: List[ShotSection] = field(default_factory=list)
    raw_content: str = ""


# ──────────────────────────── Parser ────────────────────────────

def parse_script(filepath: str) -> ParsedScript:
    with open(filepath, "r", encoding="utf-8") as f:
        content = f.read()

    result = ParsedScript(raw_content=content)

    # ── 1. Header ──
    title_m = re.search(r"^#\s+EP\d+[：:]\s*(.+)$", content, re.MULTILINE)
    if title_m:
        result.title = title_m.group(1).strip()

    meta_m = re.search(r"^>\s*(.+)$", content, re.MULTILINE)
    if meta_m:
        meta = meta_m.group(1)
        for key, attr in [("类型", "ep_type"), ("风格", "style")]:
            m = re.search(rf"{key}[：:]\s*([^｜|]+)", meta)
            if m:
                setattr(result, attr, m.group(1).strip())
        dur_m = re.search(r"时长[：:]\s*(\d+)", meta)
        if dur_m:
            result.duration = int(dur_m.group(1))
        parts = re.split(r"[｜|]", meta)
        if len(parts) > 3:
            result.extra_info = parts[-1].strip()

    # ── 2. Character table ──
    char_table_m = re.search(
        r"\|\s*角色\s*\|\s*造型\s*\|\s*出场\s*\|\s*分镜简称\s*\|\s*\n"
        r"\|[-\s]+\|[-\s]+\|[-\s]+\|[-\s]+\|\s*\n"
        r"((?:\|.+\|\s*\n)+)",
        content,
    )
    if char_table_m:
        for row in char_table_m.group(1).strip().split("\n"):
            cols = [c.strip() for c in row.split("|")[1:-1]]
            if len(cols) >= 4:
                name_raw = cols[0].strip()
                is_lead = "**" in name_raw
                name = name_raw.replace("**", "").strip()
                outfit = cols[1].strip()
                shots_str = cols[2].strip()
                brief = cols[3].strip()
                full_name = f"{name}-{outfit}" if outfit and outfit != "—" else name
                result.characters.append(
                    CharacterEntry(
                        name=name,
                        outfit=outfit,
                        shots=shots_str,
                        brief=brief,
                        full_name=full_name,
                        is_lead=is_lead,
                    )
                )

    # ── 3. Appearance details ──
    # Match **Name**：description blocks, but exclude shot-section keywords
    shot_keywords = {"首帧", "过程", "尾帧", "道具", "场景", "角色", "造型", "景别", "承接上镜"}
    app_matches = re.findall(
        r"\*\*([^*\n]+)\*\*[：:]\s*(.+?)(?=\n\n|\n\*\*|\n---|\Z)",
        content,
        re.DOTALL,
    )
    app_map: Dict[str, str] = {}
    for name, desc in app_matches:
        name = name.strip()
        if name not in shot_keywords:
            app_map[name] = desc.strip().replace("\n", " ")

    for ch in result.characters:
        if ch.full_name in app_map:
            ch.appearance = app_map[ch.full_name]

    # ── 4. Shot sections ──
    shot_re = re.compile(r"^##\s+S(\d+)\s*[·•]\s*(.+?)（(\d+)s）", re.MULTILINE)
    shot_starts = list(shot_re.finditer(content))

    for i, m in enumerate(shot_starts):
        start = m.start()
        end = shot_starts[i + 1].start() if i + 1 < len(shot_starts) else len(content)
        block = content[start:end]

        timeline_pos = block.find("## 时间轴总览")
        if timeline_pos > 0:
            block = block[:timeline_pos]

        shot = ShotSection(
            number=int(m.group(1)),
            title=m.group(2).strip(),
            duration=int(m.group(3)),
        )

        meta = re.search(
            r"-\s*\*\*场景\*\*[：:]\s*(.+?)\s*[｜|]\s*\*\*造型\*\*[：:]\s*(.+?)\s*[｜|]\s*\*\*景别\*\*[：:]\s*(.+)",
            block,
        )
        if meta:
            shot.location = meta.group(1).strip()
            shot.main_outfit = meta.group(2).strip()
            shot.shot_type = meta.group(3).strip()

        char_m = re.search(r"-\s*\*\*角色\*\*[：:]\s*(.+)", block)
        if char_m:
            raw_chars = [c.strip() for c in re.split(r"[、,，]", char_m.group(1))]
            shot.characters = [re.sub(r"[（(].+?[）)]", "", c).strip() for c in raw_chars]
        else:
            shot.characters = [shot.main_outfit]

        prev_m = re.search(r"-\s*\*\*承接上镜\*\*[：:]\s*(.+)", block)
        if prev_m:
            shot.prev_shot_ref = prev_m.group(1).strip()

        for field_name, pattern in [
            ("first_frame", r"\*\*首帧\*\*"),
            ("middle_action", r"\*\*过程\*\*"),
            ("last_frame", r"\*\*尾帧\*\*"),
        ]:
            desc_m = re.search(
                rf"{pattern}[：:]\s*(.+?)(?=\n\n|\n\*\*|\n---|\Z)",
                block,
                re.DOTALL,
            )
            if desc_m:
                setattr(shot, field_name, desc_m.group(1).strip().replace("\n", " "))

        props_m = re.search(r"\*\*道具\*\*[：:]\s*(.+)", block)
        if props_m:
            shot.props = [p.strip() for p in re.split(r"[、,，]", props_m.group(1))]

        result.shots.append(shot)

    return result


# ──────────────────────────── Importer ────────────────────────────

def import_to_db(
    script: ParsedScript,
    db_path: str,
    drama_id: int,
    episode_number: int,
    episode_title: str = "",
    episode_desc: str = "",
    dry_run: bool = False,
):
    conn = sqlite3.connect(db_path)
    conn.row_factory = sqlite3.Row
    cur = conn.cursor()
    now = datetime.utcnow().strftime("%Y-%m-%d %H:%M:%S")

    stats = {
        "episode": 0, "characters": 0, "scenes": 0,
        "props": 0, "storyboards": 0, "links": 0,
    }

    try:
        # ── Episode ──
        title = episode_title or script.title
        desc = episode_desc or None

        cur.execute(
            "SELECT id FROM episodes WHERE drama_id=? AND episode_number=? AND deleted_at IS NULL",
            (drama_id, episode_number),
        )
        ep_row = cur.fetchone()

        if ep_row:
            episode_id = ep_row["id"]
            cur.execute(
                "UPDATE episodes SET script_content=?, title=?, description=?, duration=?, updated_at=? WHERE id=?",
                (script.raw_content, title, desc, script.duration, now, episode_id),
            )
            print(f"  ✓ 更新章节 EP{episode_number:02d} (id={episode_id}): {title}")
        else:
            cur.execute(
                """INSERT INTO episodes
                   (drama_id, episode_number, title, description, duration, status, script_content, created_at, updated_at)
                   VALUES (?, ?, ?, ?, ?, 'draft', ?, ?, ?)""",
                (drama_id, episode_number, title, desc, script.duration, script.raw_content, now, now),
            )
            episode_id = cur.lastrowid
            stats["episode"] = 1
            print(f"  + 创建章节 EP{episode_number:02d} (id={episode_id}): {title}")

            cur.execute(
                "SELECT MAX(episode_number) FROM episodes WHERE drama_id=? AND deleted_at IS NULL",
                (drama_id,),
            )
            max_ep = cur.fetchone()[0] or 0
            cur.execute("UPDATE dramas SET total_episodes=?, updated_at=? WHERE id=?", (max_ep, now, drama_id))

        # ── Characters ──
        char_id_map: Dict[str, int] = {}

        for ch in script.characters:
            # Find or create base character
            cur.execute(
                "SELECT id FROM characters WHERE drama_id=? AND name=? AND parent_id IS NULL AND deleted_at IS NULL",
                (drama_id, ch.name),
            )
            base_row = cur.fetchone()

            if base_row:
                base_id = base_row["id"]
            else:
                role = "lead" if ch.is_lead else "supporting"
                cur.execute(
                    "INSERT INTO characters (drama_id, name, role, created_at, updated_at) VALUES (?, ?, ?, ?, ?)",
                    (drama_id, ch.name, role, now, now),
                )
                base_id = cur.lastrowid
                stats["characters"] += 1
                print(f"  + 基础角色: {ch.name} (id={base_id})")

            # Find or create outfit variant
            if ch.outfit and ch.outfit != "—":
                cur.execute(
                    "SELECT id FROM characters WHERE drama_id=? AND name=? AND outfit_name=? AND deleted_at IS NULL",
                    (drama_id, ch.full_name, ch.outfit),
                )
                outfit_row = cur.fetchone()

                if outfit_row:
                    char_id_map[ch.full_name] = outfit_row["id"]
                    if ch.appearance:
                        cur.execute(
                            "UPDATE characters SET appearance=?, prompt=?, updated_at=? WHERE id=?",
                            (ch.appearance, ch.brief, now, outfit_row["id"]),
                        )
                    print(f"  ✓ 角色造型: {ch.full_name} (id={outfit_row['id']})")
                else:
                    role = "lead" if ch.is_lead else "supporting"
                    cur.execute(
                        """INSERT INTO characters
                           (drama_id, name, outfit_name, parent_id, appearance, prompt, role, created_at, updated_at)
                           VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)""",
                        (drama_id, ch.full_name, ch.outfit, base_id, ch.appearance, ch.brief, role, now, now),
                    )
                    cid = cur.lastrowid
                    char_id_map[ch.full_name] = cid
                    stats["characters"] += 1
                    print(f"  + 角色造型: {ch.full_name} (id={cid})")
            else:
                # No outfit variant — use base character directly
                char_id_map[ch.full_name] = base_id
                if ch.appearance:
                    cur.execute(
                        "UPDATE characters SET appearance=?, prompt=?, updated_at=? WHERE id=?",
                        (ch.appearance, ch.brief, now, base_id),
                    )
                print(f"  ✓ 角色(无造型): {ch.full_name} (id={base_id})")

        # ── Scenes ──
        scene_id_map: Dict[str, int] = {}
        locations = sorted({s.location for s in script.shots if s.location})

        for loc in locations:
            cur.execute(
                "SELECT id FROM scenes WHERE drama_id=? AND episode_id=? AND location=? AND deleted_at IS NULL",
                (drama_id, episode_id, loc),
            )
            row = cur.fetchone()

            if row:
                scene_id_map[loc] = row["id"]
                print(f"  ✓ 场景: {loc} (id={row['id']})")
            else:
                cur.execute(
                    """INSERT INTO scenes (drama_id, episode_id, name, location, status, created_at, updated_at)
                       VALUES (?, ?, ?, ?, 'pending', ?, ?)""",
                    (drama_id, episode_id, loc, loc, now, now),
                )
                sid = cur.lastrowid
                scene_id_map[loc] = sid
                stats["scenes"] += 1
                print(f"  + 场景: {loc} (id={sid})")

        # ── Props ──
        prop_id_map: Dict[str, int] = {}
        all_props = sorted({p for s in script.shots for p in s.props})

        for pname in all_props:
            cur.execute(
                "SELECT id FROM props WHERE drama_id=? AND name=? AND deleted_at IS NULL",
                (drama_id, pname),
            )
            row = cur.fetchone()

            if row:
                prop_id_map[pname] = row["id"]
                print(f"  ✓ 道具: {pname} (id={row['id']})")
            else:
                cur.execute(
                    "INSERT INTO props (drama_id, name, created_at, updated_at) VALUES (?, ?, ?, ?)",
                    (drama_id, pname, now, now),
                )
                pid = cur.lastrowid
                prop_id_map[pname] = pid
                stats["props"] += 1
                print(f"  + 道具: {pname} (id={pid})")

        # ── Clean existing storyboards for this episode ──
        cur.execute("SELECT id FROM storyboards WHERE episode_id=? AND deleted_at IS NULL", (episode_id,))
        old_sb_ids = [r["id"] for r in cur.fetchall()]
        if old_sb_ids:
            ph = ",".join(["?"] * len(old_sb_ids))
            cur.execute(f"DELETE FROM storyboard_characters WHERE storyboard_id IN ({ph})", old_sb_ids)
            cur.execute(f"DELETE FROM storyboard_props WHERE storyboard_id IN ({ph})", old_sb_ids)
            cur.execute(f"DELETE FROM storyboards WHERE id IN ({ph})", old_sb_ids)
            print(f"  ⟳ 清理旧分镜 {len(old_sb_ids)} 条")

        # ── Storyboards ──
        for shot in script.shots:
            scene_id = scene_id_map.get(shot.location)

            cur.execute(
                """INSERT INTO storyboards
                   (episode_id, scene_id, storyboard_number, title, location, shot_type, duration,
                    first_frame_desc, middle_action_desc, last_frame_desc, prev_shot_ref,
                    status, created_at, updated_at)
                   VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'pending', ?, ?)""",
                (
                    episode_id, scene_id, shot.number, shot.title,
                    shot.location, shot.shot_type, shot.duration,
                    shot.first_frame or None,
                    shot.middle_action or None,
                    shot.last_frame or None,
                    shot.prev_shot_ref or None,
                    now, now,
                ),
            )
            sb_id = cur.lastrowid
            stats["storyboards"] += 1

            # Link characters
            for char_ref in shot.characters:
                cid = char_id_map.get(char_ref)
                if not cid:
                    cur.execute(
                        "SELECT id FROM characters WHERE drama_id=? AND name=? AND deleted_at IS NULL",
                        (drama_id, char_ref),
                    )
                    r = cur.fetchone()
                    if r:
                        cid = r["id"]
                if cid:
                    cur.execute(
                        "INSERT OR IGNORE INTO storyboard_characters (storyboard_id, character_id) VALUES (?, ?)",
                        (sb_id, cid),
                    )
                    stats["links"] += 1
                else:
                    print(f"  ⚠ S{shot.number:02d}: 角色 '{char_ref}' 未找到匹配")

            # Link props
            for pref in shot.props:
                pid = prop_id_map.get(pref)
                if pid:
                    cur.execute(
                        "INSERT OR IGNORE INTO storyboard_props (storyboard_id, prop_id) VALUES (?, ?)",
                        (sb_id, pid),
                    )
                    stats["links"] += 1
                else:
                    print(f"  ⚠ S{shot.number:02d}: 道具 '{pref}' 未找到匹配")

            print(
                f"  + 分镜 S{shot.number:02d} · {shot.title} ({shot.duration}s) "
                f"@ {shot.location} [角色:{len(shot.characters)} 道具:{len(shot.props)}]"
            )

        # ── Episode ↔ Character links ──
        cur.execute("DELETE FROM episode_characters WHERE episode_id=?", (episode_id,))
        for cid in set(char_id_map.values()):
            cur.execute(
                "INSERT OR IGNORE INTO episode_characters (episode_id, character_id) VALUES (?, ?)",
                (episode_id, cid),
            )
            stats["links"] += 1

        # ── Episode ↔ Prop links ──
        cur.execute("DELETE FROM episode_props WHERE episode_id=?", (episode_id,))
        for pid in prop_id_map.values():
            cur.execute(
                "INSERT OR IGNORE INTO episode_props (episode_id, prop_id) VALUES (?, ?)",
                (episode_id, pid),
            )
            stats["links"] += 1

        if dry_run:
            print("\n  [DRY RUN] 回滚所有变更")
            conn.rollback()
        else:
            conn.commit()
            print(f"\n  ✅ 导入完成!")

        print(
            f"  统计: 章节={stats['episode']}, 角色={stats['characters']}, "
            f"场景={stats['scenes']}, 道具={stats['props']}, "
            f"分镜={stats['storyboards']}, 关联={stats['links']}"
        )

    except Exception as e:
        conn.rollback()
        print(f"\n  ❌ 导入失败: {e}")
        import traceback
        traceback.print_exc()
        sys.exit(1)
    finally:
        conn.close()


# ──────────────────────────── CLI ────────────────────────────

def extract_episode_number(filepath: str) -> Optional[int]:
    m = re.search(r"EP(\d+)", os.path.basename(filepath))
    return int(m.group(1)) if m else None


def main():
    parser = argparse.ArgumentParser(
        description="CineMaker 剧本自动导入工具",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""示例:
  python3 scripts/import_script.py docs/ai-vlogs/都市生活/EP03-xxx.md --drama-id 13
  python3 scripts/import_script.py EP03.md --drama-id 13 --title "自定义标题" --desc "简介"
  python3 scripts/import_script.py EP03.md --drama-id 13 --dry-run
  python3 scripts/import_script.py EP03.md --drama-id 13 --parse-only
""",
    )
    parser.add_argument("script", help="剧本 Markdown 文件路径")
    parser.add_argument("--drama-id", type=int, required=True, help="剧集 ID")
    parser.add_argument("--episode", type=int, help="章节号（默认从文件名提取）")
    parser.add_argument("--title", help="自定义章节标题（默认从剧本提取）")
    parser.add_argument("--desc", help="章节简介")
    parser.add_argument("--db", default="data/drama_generator.db", help="数据库路径 (默认: data/drama_generator.db)")
    parser.add_argument("--dry-run", action="store_true", help="预览模式，不写入数据库")
    parser.add_argument("--parse-only", action="store_true", help="仅解析剧本，不导入")

    args = parser.parse_args()

    if not os.path.exists(args.script):
        print(f"错误: 文件不存在: {args.script}")
        sys.exit(1)

    ep_num = args.episode or extract_episode_number(args.script)
    if not ep_num and not args.parse_only:
        print("错误: 无法从文件名提取章节号，请用 --episode 指定")
        sys.exit(1)

    # ── Parse ──
    print(f"📖 解析剧本: {args.script}")
    script = parse_script(args.script)

    print(f"  标题: {script.title}")
    print(f"  类型: {script.ep_type} | 时长: {script.duration}s | 风格: {script.style}")
    print(f"  角色: {len(script.characters)} 个")
    for ch in script.characters:
        lead_tag = "★" if ch.is_lead else " "
        app_tag = "✓" if ch.appearance else "✗"
        print(f"    {lead_tag}[{app_tag}] {ch.full_name} → {ch.shots}")
    print(f"  镜头: {len(script.shots)} 个")
    total_dur = 0
    for s in script.shots:
        total_dur += s.duration
        props_str = f"  道具:{','.join(s.props)}" if s.props else ""
        print(f"    S{s.number:02d} · {s.title} ({s.duration}s) @ {s.location} [{', '.join(s.characters)}]{props_str}")
    print(f"  总时长: {total_dur}s (标注: {script.duration}s)")

    if args.parse_only:
        return

    if not os.path.exists(args.db):
        print(f"错误: 数据库不存在: {args.db}")
        sys.exit(1)

    # ── Import ──
    mode = "🔍 预览模式" if args.dry_run else "📥 开始导入"
    print(f"\n{mode}")
    print(f"  数据库: {args.db}")
    print(f"  剧集 ID: {args.drama_id}")
    print(f"  章节号: EP{ep_num:02d}\n")

    import_to_db(
        script=script,
        db_path=args.db,
        drama_id=args.drama_id,
        episode_number=ep_num,
        episode_title=args.title or "",
        episode_desc=args.desc or "",
        dry_run=args.dry_run,
    )


if __name__ == "__main__":
    main()
