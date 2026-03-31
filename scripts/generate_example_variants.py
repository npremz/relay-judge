#!/usr/bin/env python3

from __future__ import annotations

from pathlib import Path


ROOT = Path(__file__).resolve().parent.parent
VARIANTS_DIR = ROOT / "examples" / "variants"


SUBJECTS = [
    {
        "file_name": "two_sum.py",
        "wrong": """def two_sum(nums, target):\n    return [0, 0]\n""",
        "runtime": """def two_sum(nums, target):\n    raise RuntimeError("intentional runtime error")\n""",
        "syntax": """def two_sum(nums, target)\n    return []\n""",
        "timeout": """import time\n\n\ndef two_sum(nums, target):\n    time.sleep(6)\n    return []\n""",
        "slow": """def two_sum(nums, target):\n    for i in range(len(nums)):\n        for j in range(i + 1, len(nums)):\n            if nums[i] + nums[j] == target:\n                return [i, j]\n    return []\n""",
    },
    {
        "file_name": "minimum_size_subarray_sum.py",
        "wrong": """def minimum_size_subarray_sum(target, nums):\n    return 0\n""",
        "runtime": """def minimum_size_subarray_sum(target, nums):\n    raise RuntimeError("intentional runtime error")\n""",
        "syntax": """def minimum_size_subarray_sum(target, nums)\n    return 0\n""",
        "timeout": """import time\n\n\ndef minimum_size_subarray_sum(target, nums):\n    time.sleep(6)\n    return 0\n""",
        "slow": """def minimum_size_subarray_sum(target, nums):\n    best = len(nums) + 1\n    for start in range(len(nums)):\n        total = 0\n        for end in range(start, len(nums)):\n            total += nums[end]\n            if total >= target:\n                best = min(best, end - start + 1)\n                break\n    return 0 if best == len(nums) + 1 else best\n""",
    },
    {
        "file_name": "first_unique_character.py",
        "wrong": """def first_unique_character(s):\n    return -1\n""",
        "runtime": """def first_unique_character(s):\n    raise RuntimeError("intentional runtime error")\n""",
        "syntax": """def first_unique_character(s)\n    return -1\n""",
        "timeout": """import time\n\n\ndef first_unique_character(s):\n    time.sleep(6)\n    return -1\n""",
        "slow": """def first_unique_character(s):\n    for index, char in enumerate(s):\n        if s.count(char) == 1:\n            return index\n    return -1\n""",
    },
    {
        "file_name": "longest_substring_without_repeating_characters.py",
        "wrong": """def longest_substring_without_repeating_characters(s):\n    return 1 if s else 0\n""",
        "runtime": """def longest_substring_without_repeating_characters(s):\n    raise RuntimeError("intentional runtime error")\n""",
        "syntax": """def longest_substring_without_repeating_characters(s)\n    return 0\n""",
        "timeout": """import time\n\n\ndef longest_substring_without_repeating_characters(s):\n    time.sleep(6)\n    return 0\n""",
        "slow": """def longest_substring_without_repeating_characters(s):\n    best = 0\n    for start in range(len(s)):\n        seen = set()\n        for end in range(start, len(s)):\n            char = s[end]\n            if char in seen:\n                break\n            seen.add(char)\n            best = max(best, end - start + 1)\n    return best\n""",
    },
    {
        "file_name": "merge_intervals.py",
        "wrong": """def merge_intervals(intervals):\n    return intervals\n""",
        "runtime": """def merge_intervals(intervals):\n    raise RuntimeError("intentional runtime error")\n""",
        "syntax": """def merge_intervals(intervals)\n    return []\n""",
        "timeout": """import time\n\n\ndef merge_intervals(intervals):\n    time.sleep(6)\n    return intervals\n""",
        "slow": """def merge_intervals(intervals):\n    if not intervals:\n        return []\n\n    pending = [interval[:] for interval in intervals]\n    changed = True\n    while changed:\n        changed = False\n        next_pending = []\n        while pending:\n            current = pending.pop()\n            merged = False\n            for index, other in enumerate(pending):\n                if current[0] <= other[1] and other[0] <= current[1]:\n                    pending[index] = [min(current[0], other[0]), max(current[1], other[1])]\n                    changed = True\n                    merged = True\n                    break\n            if not merged:\n                next_pending.append(current)\n        pending = next_pending\n    return sorted(pending)\n""",
    },
    {
        "file_name": "top_k_frequent_elements.py",
        "wrong": """def top_k_frequent_elements(nums, k):\n    return nums[:k]\n""",
        "runtime": """def top_k_frequent_elements(nums, k):\n    raise RuntimeError("intentional runtime error")\n""",
        "syntax": """def top_k_frequent_elements(nums, k)\n    return []\n""",
        "timeout": """import time\n\n\ndef top_k_frequent_elements(nums, k):\n    time.sleep(6)\n    return []\n""",
        "slow": """def top_k_frequent_elements(nums, k):\n    ordered = []\n    for value in set(nums):\n        ordered.append((nums.count(value), value))\n    ordered.sort(reverse=True)\n    return [value for _count, value in ordered[:k]]\n""",
    },
    {
        "file_name": "valid_parentheses.py",
        "wrong": """def valid_parentheses(s):\n    return True\n""",
        "runtime": """def valid_parentheses(s):\n    raise RuntimeError("intentional runtime error")\n""",
        "syntax": """def valid_parentheses(s)\n    return False\n""",
        "timeout": """import time\n\n\ndef valid_parentheses(s):\n    time.sleep(6)\n    return False\n""",
        "slow": """def valid_parentheses(s):\n    previous = None\n    current = s\n    while previous != current:\n        previous = current\n        current = current.replace("()", "").replace("[]", "").replace("{}", "")\n    return current == ""\n""",
    },
]


def generate() -> None:
    for variant in ["slow", "wrong", "runtime", "syntax", "timeout"]:
        target_dir = VARIANTS_DIR / variant
        target_dir.mkdir(parents=True, exist_ok=True)
        for subject in SUBJECTS:
            path = target_dir / subject["file_name"]
            path.write_text(subject[variant], encoding="utf-8")


def main() -> int:
    generate()
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
