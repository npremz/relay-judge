def top_k_frequent_elements(nums, k):
    ordered = []
    for value in set(nums):
        ordered.append((nums.count(value), value))
    ordered.sort(reverse=True)
    return [value for _count, value in ordered[:k]]
