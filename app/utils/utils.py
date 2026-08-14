import math

def format_bytes(size_bytes: int) -> str:
    if size_bytes == -1:
        return "∞"

    if size_bytes < 0:
        return "0 Б"

    labels = ["Б", "КБ", "МБ", "ГБ", "ТБ"]
    if size_bytes == 0:
        return "0 Б"


    i = int(math.floor(math.log(size_bytes, 1024)))
    if i >= len(labels):
        i = len(labels) - 1

    p = math.pow(1024, i)
    s = round(size_bytes / p, 2)

    if s.is_integer():
        s = int(s)

    return f"{s} {labels[i]}"
