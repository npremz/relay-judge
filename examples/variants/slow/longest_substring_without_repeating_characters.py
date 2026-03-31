def longest_substring_without_repeating_characters(s):
    best = 0
    for start in range(len(s)):
        seen = set()
        for end in range(start, len(s)):
            char = s[end]
            if char in seen:
                break
            seen.add(char)
            best = max(best, end - start + 1)
    return best
