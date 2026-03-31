def valid_parentheses(s):
    pairs = {")": "(", "]": "[", "}": "{"}
    stack = []

    for char in s:
        if char in pairs.values():
            stack.append(char)
            continue

        if not stack or stack[-1] != pairs.get(char):
            return False

        stack.pop()

    return len(stack) == 0
