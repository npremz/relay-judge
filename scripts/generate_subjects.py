#!/usr/bin/env python3

from __future__ import annotations

import json
from pathlib import Path


ROOT = Path(__file__).resolve().parent.parent
SUBJECTS_DIR = ROOT / "subjects"


def write_subject(spec: dict) -> None:
    path = SUBJECTS_DIR / spec["id"] / "subject.json"
    path.write_text(
        json.dumps(spec, indent=2, ensure_ascii=True) + "\n",
        encoding="utf-8",
    )


def unique_unicode_run(length: int, start: int = 0x0100) -> str:
    chars: list[str] = []
    codepoint = start

    while len(chars) < length:
        if 0xD800 <= codepoint <= 0xDFFF:
            codepoint = 0xE000
        chars.append(chr(codepoint))
        codepoint += 1

    return "".join(chars)


def build_top_k_perf_case(distinct: int = 256, max_frequency: int = 512) -> tuple[list[int], list[int]]:
    nums: list[int] = []
    for value in range(distinct):
        nums.extend([value] * (max_frequency - value))
    return nums, list(range(8))


def build_merge_intervals_perf_case(interval_count: int = 12000) -> tuple[list[list[int]], list[list[int]]]:
    intervals = [[start, start + 1] for start in range(interval_count - 1, -1, -1)]
    return intervals, [[0, interval_count]]


def build_specs() -> list[dict]:
    longest_unique_run = unique_unicode_run(4096)
    top_k_perf_nums, top_k_perf_expected = build_top_k_perf_case()
    merge_perf_intervals, merge_perf_expected = build_merge_intervals_perf_case()

    return [
        {
            "id": "two-sum",
            "title": "Two Sum",
            "prototype": "def two_sum(nums, target):",
            "difficulty": "easy",
            "description": (
                "Étant donné un tableau d'entiers `nums` et un entier `target`, retourne les indices des deux "
                "éléments dont la somme vaut exactement `target`.\n\n"
                "Tu peux supposer qu'il existe une solution unique et qu'un même élément ne peut pas être utilisé "
                "deux fois.\n\n"
                "Le résultat peut être retourné dans n'importe quel ordre.\n\n"
                "Exemple 1:\n"
                "nums = [2, 7, 11, 15], target = 9\n"
                "Réponse attendue: [0, 1]\n"
                "Explication: nums[0] + nums[1] = 2 + 7 = 9\n\n"
                "Exemple 2:\n"
                "nums = [3, 3], target = 6\n"
                "Réponse attendue: [0, 1]\n\n"
                "Attendu:\n"
                "- gérer les doublons\n"
                "- fonctionner avec des nombres négatifs\n"
                "- éviter une solution fragile basée uniquement sur les exemples"
            ),
            "file_name": "two_sum.py",
            "function_name": "two_sum",
            "checker": "two_sum_pair",
            "time_limit_ms": 1500,
            "tests": [
                {
                    "name": "basic_pair",
                    "group": "core",
                    "args": [[2, 7, 11, 15], 9],
                },
                {
                    "name": "duplicate_values",
                    "group": "core",
                    "args": [[3, 3], 6],
                },
                {
                    "name": "late_solution",
                    "group": "edge",
                    "args": [[5, 17, 1, 8, 12, 20], 21],
                },
                {
                    "name": "negative_numbers",
                    "group": "edge",
                    "args": [[-4, 10, 7, 15, -1], 6],
                },
                {
                    "name": "anti_hardcode_reordered_values",
                    "group": "anti-hardcode",
                    "args": [[10, 1, 9, 4, 6], 10],
                },
                {
                    "name": "large_perf_case",
                    "group": "perf",
                    "args": [[10 * index for index in range(40000)] + [1, 3], 4],
                },
            ],
        },
        {
            "id": "minimum-size-subarray-sum",
            "title": "Minimum Size Subarray Sum",
            "prototype": "def minimum_size_subarray_sum(target, nums):",
            "difficulty": "medium",
            "description": (
                "Étant donné un tableau d'entiers positifs `nums` et un entier positif `target`, retourne la "
                "longueur minimale d'une sous-suite contiguë dont la somme est supérieure ou égale à `target`.\n\n"
                "S'il n'existe aucune sous-suite valide, retourne `0`.\n\n"
                "Exemple 1:\n"
                "target = 7, nums = [2, 3, 1, 2, 4, 3]\n"
                "Réponse attendue: 2\n"
                "Explication: la sous-suite [4, 3] atteint 7 avec la longueur minimale.\n\n"
                "Exemple 2:\n"
                "target = 4, nums = [1, 4, 4]\n"
                "Réponse attendue: 1\n\n"
                "Exemple 3:\n"
                "target = 100, nums = [1, 2, 3, 4, 5]\n"
                "Réponse attendue: 0\n\n"
                "Attendu:\n"
                "- bien gérer les fenêtres qui doivent s'agrandir puis se réduire plusieurs fois\n"
                "- retourner 0 si aucune solution n'existe"
            ),
            "file_name": "minimum_size_subarray_sum.py",
            "function_name": "minimum_size_subarray_sum",
            "checker": "exact_int",
            "time_limit_ms": 1500,
            "tests": [
                {
                    "name": "leetcode_example_1",
                    "group": "core",
                    "args": [7, [2, 3, 1, 2, 4, 3]],
                    "expected": 2,
                },
                {
                    "name": "single_element_match",
                    "group": "core",
                    "args": [4, [1, 4, 4]],
                    "expected": 1,
                },
                {
                    "name": "no_solution",
                    "group": "edge",
                    "args": [100, [1, 2, 3, 4, 5]],
                    "expected": 0,
                },
                {
                    "name": "window_moves_multiple_times",
                    "group": "edge",
                    "args": [11, [1, 2, 3, 4, 5]],
                    "expected": 3,
                },
                {
                    "name": "anti_hardcode_short_best_window",
                    "group": "anti-hardcode",
                    "args": [8, [2, 1, 5, 2, 3, 2]],
                    "expected": 3,
                },
                {
                    "name": "perf_long_prefix",
                    "group": "perf",
                    "args": [20000, [1] * 40000],
                    "expected": 20000,
                },
            ],
        },
        {
            "id": "first-unique-character",
            "title": "First Unique Character in a String",
            "prototype": "def first_unique_character(s):",
            "difficulty": "easy",
            "description": (
                "Étant donnée une chaîne `s`, retourne l'indice du premier caractère qui n'apparaît qu'une seule "
                "fois.\n\n"
                "Si aucun caractère n'est unique, retourne `-1`.\n\n"
                "L'indice est basé sur 0.\n\n"
                "Exemple 1:\n"
                "s = \"leetcode\"\n"
                "Réponse attendue: 0\n"
                "Explication: `l` est le premier caractère non répété.\n\n"
                "Exemple 2:\n"
                "s = \"loveleetcode\"\n"
                "Réponse attendue: 2\n"
                "Explication: `v` est le premier caractère présent une seule fois.\n\n"
                "Exemple 3:\n"
                "s = \"aabbcc\"\n"
                "Réponse attendue: -1\n\n"
                "Attendu:\n"
                "- gérer le cas où le caractère unique est à la fin\n"
                "- gérer le cas où aucun caractère n'est valide\n"
                "- éviter une solution trop coûteuse sur de longues chaînes"
            ),
            "file_name": "first_unique_character.py",
            "function_name": "first_unique_character",
            "checker": "exact_int",
            "time_limit_ms": 1200,
            "tests": [
                {
                    "name": "leetcode_example_1",
                    "group": "core",
                    "args": ["leetcode"],
                    "expected": 0,
                },
                {
                    "name": "leetcode_example_2",
                    "group": "core",
                    "args": ["loveleetcode"],
                    "expected": 2,
                },
                {
                    "name": "no_unique_character",
                    "group": "edge",
                    "args": ["aabbcc"],
                    "expected": -1,
                },
                {
                    "name": "unique_in_middle",
                    "group": "edge",
                    "args": ["aabbcddef"],
                    "expected": 4,
                },
                {
                    "name": "anti_hardcode_unique_at_end",
                    "group": "anti-hardcode",
                    "args": ["zzxxyywwv"],
                    "expected": 8,
                },
                {
                    "name": "long_perf_case",
                    "group": "perf",
                    "args": [("a" * 50000) + "b"],
                    "expected": 50000,
                },
            ],
        },
        {
            "id": "longest-substring-without-repeating-characters",
            "title": "Longest Substring Without Repeating Characters",
            "prototype": "def longest_substring_without_repeating_characters(s):",
            "difficulty": "medium",
            "description": (
                "Étant donnée une chaîne `s`, retourne la longueur de la plus longue sous-chaîne contenant "
                "uniquement des caractères distincts.\n\n"
                "Une sous-chaîne est une portion contiguë de la chaîne.\n\n"
                "Exemple 1:\n"
                "s = \"abcabcbb\"\n"
                "Réponse attendue: 3\n"
                "Explication: la plus longue sous-chaîne sans répétition est \"abc\".\n\n"
                "Exemple 2:\n"
                "s = \"bbbbb\"\n"
                "Réponse attendue: 1\n"
                "Explication: la meilleure sous-chaîne possible est \"b\".\n\n"
                "Exemple 3:\n"
                "s = \"pwwkew\"\n"
                "Réponse attendue: 3\n"
                "Explication: \"wke\" est valide, mais \"pwke\" n'est pas une sous-chaîne contiguë.\n\n"
                "Exemple 4:\n"
                "s = \"\"\n"
                "Réponse attendue: 0"
            ),
            "file_name": "longest_substring_without_repeating_characters.py",
            "function_name": "longest_substring_without_repeating_characters",
            "checker": "exact_int",
            "time_limit_ms": 1500,
            "tests": [
                {
                    "name": "leetcode_example_1",
                    "group": "core",
                    "args": ["abcabcbb"],
                    "expected": 3,
                },
                {
                    "name": "all_same",
                    "group": "core",
                    "args": ["bbbbb"],
                    "expected": 1,
                },
                {
                    "name": "overlapping_window",
                    "group": "edge",
                    "args": ["pwwkew"],
                    "expected": 3,
                },
                {
                    "name": "empty_string",
                    "group": "edge",
                    "args": [""],
                    "expected": 0,
                },
                {
                    "name": "anti_hardcode_repeated_after_gap",
                    "group": "anti-hardcode",
                    "args": ["dvdf"],
                    "expected": 3,
                },
                {
                    "name": "perf_unique_tail",
                    "group": "perf",
                    "args": [longest_unique_run + longest_unique_run[0]],
                    "expected": len(longest_unique_run),
                },
            ],
        },
        {
            "id": "merge-intervals",
            "title": "Merge Intervals",
            "prototype": "def merge_intervals(intervals):",
            "difficulty": "medium",
            "description": (
                "Étant donnée une liste d'intervalles `intervals` où chaque intervalle est de la forme `[start, end]`, "
                "fusionne tous les intervalles qui se chevauchent et retourne la liste finale normalisée.\n\n"
                "Deux intervalles qui se touchent doivent aussi être fusionnés si la fin de l'un est égale au début "
                "de l'autre.\n\n"
                "Le résultat doit couvrir exactement les mêmes plages que l'entrée, sans doublons ni recouvrements.\n\n"
                "Exemple 1:\n"
                "intervals = [[1,3],[2,6],[8,10],[15,18]]\n"
                "Réponse attendue: [[1,6],[8,10],[15,18]]\n\n"
                "Exemple 2:\n"
                "intervals = [[1,4],[4,5]]\n"
                "Réponse attendue: [[1,5]]\n\n"
                "Exemple 3:\n"
                "intervals = [[6,8],[1,9],[2,4],[4,7]]\n"
                "Réponse attendue: [[1,9]]\n\n"
                "Attendu:\n"
                "- accepter une entrée non triée\n"
                "- gérer les intervalles imbriqués\n"
                "- produire une sortie proprement ordonnée"
            ),
            "file_name": "merge_intervals.py",
            "function_name": "merge_intervals",
            "checker": "intervals_exact",
            "time_limit_ms": 1500,
            "tests": [
                {
                    "name": "simple_overlap",
                    "group": "core",
                    "args": [[[1, 3], [2, 6], [8, 10], [15, 18]]],
                    "expected": [[1, 6], [8, 10], [15, 18]],
                },
                {
                    "name": "touching_intervals",
                    "group": "core",
                    "args": [[[1, 4], [4, 5]]],
                    "expected": [[1, 5]],
                },
                {
                    "name": "already_sorted_with_gaps",
                    "group": "edge",
                    "args": [[[1, 2], [5, 7], [9, 12]]],
                    "expected": [[1, 2], [5, 7], [9, 12]],
                },
                {
                    "name": "unsorted_nested",
                    "group": "edge",
                    "args": [[[6, 8], [1, 9], [2, 4], [4, 7]]],
                    "expected": [[1, 9]],
                },
                {
                    "name": "anti_hardcode_reverse_input",
                    "group": "anti-hardcode",
                    "args": [[[9, 10], [3, 5], [4, 8], [1, 2]]],
                    "expected": [[1, 2], [3, 8], [9, 10]],
                },
                {
                    "name": "perf_many_small_intervals",
                    "group": "perf",
                    "args": [merge_perf_intervals],
                    "expected": merge_perf_expected,
                },
            ],
        },
        {
            "id": "top-k-frequent-elements",
            "title": "Top K Frequent Elements",
            "prototype": "def top_k_frequent_elements(nums, k):",
            "difficulty": "medium",
            "description": (
                "Étant donné un tableau d'entiers `nums` et un entier `k`, retourne les `k` éléments les plus "
                "fréquents du tableau.\n\n"
                "L'ordre des éléments retournés n'a pas d'importance.\n\n"
                "Exemple 1:\n"
                "nums = [1, 1, 1, 2, 2, 3], k = 2\n"
                "Réponse attendue: [1, 2]\n\n"
                "Exemple 2:\n"
                "nums = [1], k = 1\n"
                "Réponse attendue: [1]\n\n"
                "Exemple 3:\n"
                "nums = [4, -1, -1, 2, -1, 2, 3, 3, 3], k = 2\n"
                "Réponse attendue: [-1, 3]\n\n"
                "Attendu:\n"
                "- gérer des valeurs négatives\n"
                "- gérer les doublons massifs\n"
                "- ne pas dépendre d'un ordre de sortie particulier tant que le bon ensemble est renvoyé"
            ),
            "file_name": "top_k_frequent_elements.py",
            "function_name": "top_k_frequent_elements",
            "checker": "set_of_ints",
            "time_limit_ms": 1500,
            "tests": [
                {
                    "name": "leetcode_example_1",
                    "group": "core",
                    "args": [[1, 1, 1, 2, 2, 3], 2],
                    "expected": [1, 2],
                },
                {
                    "name": "single_value",
                    "group": "core",
                    "args": [[1], 1],
                    "expected": [1],
                },
                {
                    "name": "negative_values",
                    "group": "edge",
                    "args": [[4, -1, -1, 2, -1, 2, 3, 3, 3], 2],
                    "expected": [-1, 3],
                },
                {
                    "name": "k_equals_distinct_count",
                    "group": "edge",
                    "args": [[5, 5, 6, 6, 7, 8], 4],
                    "expected": [5, 6, 7, 8],
                },
                {
                    "name": "anti_hardcode_frequency_order_irrelevant",
                    "group": "anti-hardcode",
                    "args": [[9, 9, 8, 8, 8, 7, 7, 6], 2],
                    "expected": [8, 9],
                },
                {
                    "name": "perf_repeated_blocks",
                    "group": "perf",
                    "args": [top_k_perf_nums, 8],
                    "expected": top_k_perf_expected,
                },
            ],
        },
        {
            "id": "valid-parentheses",
            "title": "Valid Parentheses",
            "prototype": "def valid_parentheses(s):",
            "difficulty": "easy",
            "description": (
                "Étant donnée une chaîne `s` ne contenant que les caractères `(`, `)`, `[`, `]`, `{` et `}`, "
                "retourne `true` si la chaîne est valide, sinon `false`.\n\n"
                "Une chaîne est valide si:\n"
                "- chaque parenthèse ouvrante est fermée par le bon type de parenthèse\n"
                "- les fermetures respectent le bon ordre\n"
                "- aucune parenthèse ouvrante ne reste sans fermeture\n\n"
                "Exemple 1:\n"
                "s = \"()[]{}\"\n"
                "Réponse attendue: true\n\n"
                "Exemple 2:\n"
                "s = \"(]\"\n"
                "Réponse attendue: false\n\n"
                "Exemple 3:\n"
                "s = \"([)]\"\n"
                "Réponse attendue: false\n\n"
                "Exemple 4:\n"
                "s = \"{[()()]}\"\n"
                "Réponse attendue: true"
            ),
            "file_name": "valid_parentheses.py",
            "function_name": "valid_parentheses",
            "checker": "exact_bool",
            "time_limit_ms": 1200,
            "tests": [
                {
                    "name": "simple_valid",
                    "group": "core",
                    "args": ["()[]{}"],
                    "expected": True,
                },
                {
                    "name": "simple_invalid",
                    "group": "core",
                    "args": ["(]"],
                    "expected": False,
                },
                {
                    "name": "nested_valid",
                    "group": "edge",
                    "args": ["{[()()]}"],
                    "expected": True,
                },
                {
                    "name": "wrong_order",
                    "group": "edge",
                    "args": ["([)]"],
                    "expected": False,
                },
                {
                    "name": "unfinished_stack",
                    "group": "edge",
                    "args": ["(((()"],
                    "expected": False,
                },
                {
                    "name": "anti_hardcode_mixed_valid",
                    "group": "anti-hardcode",
                    "args": ["([]{})[{}]"],
                    "expected": True,
                },
                {
                    "name": "long_balanced",
                    "group": "perf",
                    "args": [("(" * 30000) + (")" * 30000)],
                    "expected": True,
                },
            ],
        },
    ]


def main() -> int:
    for spec in build_specs():
        write_subject(spec)

    return 0


if __name__ == "__main__":
    raise SystemExit(main())
