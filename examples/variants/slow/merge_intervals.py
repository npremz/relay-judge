def merge_intervals(intervals):
    if not intervals:
        return []

    pending = [interval[:] for interval in intervals]
    changed = True
    while changed:
        changed = False
        next_pending = []
        while pending:
            current = pending.pop()
            merged = False
            for index, other in enumerate(pending):
                if current[0] <= other[1] and other[0] <= current[1]:
                    pending[index] = [min(current[0], other[0]), max(current[1], other[1])]
                    changed = True
                    merged = True
                    break
            if not merged:
                next_pending.append(current)
        pending = next_pending
    return sorted(pending)
