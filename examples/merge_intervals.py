def merge_intervals(intervals):
    if not intervals:
        return []

    intervals = sorted(intervals)
    merged = [intervals[0][:]]

    for start, end in intervals[1:]:
        last = merged[-1]
        if start <= last[1]:
            last[1] = max(last[1], end)
        else:
            merged.append([start, end])

    return merged
