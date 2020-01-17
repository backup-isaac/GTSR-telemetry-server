import numpy as np

# Returns start/end pairs to merge
def checkTimes(start,end,merged):

    new_start = -1
    new_end = -1

    # Edge cases
    if start >= end:
        return "Invalid Times"

    if start < merged[0][0] or start > merged[-1][1]:
        new_start = start

    if end < merged[0][0] or end > merged[-1][1]:
        new_end = end

    # Get overall time range
    for pair in merged:

        # Check start if still in initial state
        if new_start == -1:
            if start < pair[0]:
                new_start = start
            elif start == pair[0] or start <= pair[1]:
                new_start = pair[1] + 1

        # Check end if still in initial state
        if new_end == -1:
            if end < pair[0]:
                new_end = end
            elif end == pair[0] or end <= pair[1]:
                new_end = pair[0] - 1


        # Check if new_start and new_end were found
        if new_start != -1 and new_end != -1:

            # Check if new_start and new_end were found in same pair
            if new_start > new_end:
                return "Times already merged"
            break

    # Remove any internal pairs from overall range
    pairs = []
    p1 = new_start
    p2 = new_end
    for pair in merged:

        if pair[1] < new_start:
            continue
        if pair[0] > new_end:
            pairs.append((p1,p2))
            break

        pairs.append((p1,pair[0]-1))
        p1 = pair[1]+1

    # Edge case
    if merged[-1][1] < new_end:
        pairs.append((p1,p2))

    # Return pairs to be merged
    return pairs

# How to use
merged = [(5,10),(15,20),(25,30),(35,40)]

start = 0
end = 14

to_merge = checkTimes(start,end,merged)

print(start,end)
print(merged)
print(to_merge)
