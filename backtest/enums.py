from enum import Enum

class Interval(Enum):
    M1 : str = "1m"
    M5 : str = "5m"
    M15 : str = "15m"
    M30 : str = "30m"
    H1 : str = "1h"
    H4 : str = "4h"
    D1 : str = "1d"
    W1 : str = "1w"
    M1_: str = "1M"