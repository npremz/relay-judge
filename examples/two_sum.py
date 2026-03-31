def two_sum(nums, target):
    seen = {}
    for index, value in enumerate(nums):
        wanted = target - value
        if wanted in seen:
            return [seen[wanted], index]
        seen[value] = index
    return []
