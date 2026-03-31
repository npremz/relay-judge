def valid_parentheses(s):
    previous = None
    current = s
    while previous != current:
        previous = current
        current = current.replace("()", "").replace("[]", "").replace("{}", "")
    return current == ""
