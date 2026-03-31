def first_unique_character(s):
    for index, char in enumerate(s):
        if s.count(char) == 1:
            return index
    return -1
