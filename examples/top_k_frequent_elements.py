def top_k_frequent_elements(nums, k):
    counts = {}
    for value in nums:
        counts[value] = counts.get(value, 0) + 1

    ordered = sorted(counts.items(), key=lambda item: item[1], reverse=True)
    return [value for value, _count in ordered[:k]]
