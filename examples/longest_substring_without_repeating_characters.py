def longest_substring_without_repeating_characters(s):
    seen = {}
    left = 0
    best = 0

    for right, char in enumerate(s):
        if char in seen and seen[char] >= left:
            left = seen[char] + 1
        seen[char] = right
        best = max(best, right - left + 1)

    return best
