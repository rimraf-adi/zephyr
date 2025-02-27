# import pandas as pd
# import numpy as np
# from pathlib import Path

# def bollinger_bands_strategy(df: pd.DataFrame, output_path: str = "../database/eurusd/processed_csv/rsi_bb.csv") -> pd.DataFrame:
#     df.columns = ["open_time", "open", "high", "low", "close", "volume"]
#     df = df.drop(columns=["volume"])  # Drop volume column

#     # Convert timestamp to datetime
#     df["open_time"] = pd.to_datetime(df["open_time"], unit="ms")

#     # Convert numeric columns to float
#     numeric_cols = ["open", "high", "low", "close"]
#     df[numeric_cols] = df[numeric_cols].astype(float)

#     # --- Calculate Bollinger Bands ---
#     window, std_dev = 20, 2
#     df["SMA"] = df["close"].rolling(window=window).mean()
#     df["Upper_BB"] = df["SMA"] + (df["close"].rolling(window=window).std() * std_dev)
#     df["Lower_BB"] = df["SMA"] - (df["close"].rolling(window=window).std() * std_dev)

#     # --- Calculate RSI ---
#     period = 14
#     delta = df["close"].diff(1)
#     gain = np.where(delta > 0, delta, 0)
#     loss = np.where(delta < 0, -delta, 0)

#     avg_gain = pd.Series(gain).ewm(span=period, adjust=False).mean()
#     avg_loss = pd.Series(loss).ewm(span=period, adjust=False).mean()

#     rs = avg_gain / avg_loss
#     df["RSI"] = 100 - (100 / (1 + rs))

#     # --- Identify Trade Signals ---
#     df["Long_Entry"] = (df["RSI"] <= 30) & (df["close"] < df["Lower_BB"])
#     df["Short_Entry"] = (df["RSI"] > 70) & (df["close"] > df["Upper_BB"])
#     df["Long_Exit"] = df["close"] >= df["Upper_BB"]
#     df["Short_Exit"] = df["close"] <= df["Lower_BB"]

#     # Convert timestamps to seconds
#     df["time"] = df["open_time"].astype("int64") // 10**9

#     # Remove duplicates and sort
#     df = df.drop_duplicates(subset=["time"]).sort_values(by="time", ascending=True)
#     df["Upper_BB"] = df["Upper_BB"].ffill()
#     df["Lower_BB"] = df["Lower_BB"].ffill()
#     df["RSI"] = df["RSI"].ffill()

#     # Export CSV
#     output_file = Path(output_path)
#     output_file.parent.mkdir(parents=True, exist_ok=True)
#     df.to_csv(output_file, index=False)

#     return df

# # Load Data
# eurusd = pd.read_excel("../database/eurusd/eurusd2022.xlsx")

# # Run Strategy
# bollingerbands = bollinger_bands_strategy(eurusd)


# import pandas as pd
# import numpy as np

# def bollinger_bands_strategy(df: pd.DataFrame) -> pd.DataFrame:
#     df.columns = ["open_time", "open", "high", "low", "close", "volume"]
#     df = df.drop(columns=["volume"])  # Drop volume column

#     # Convert timestamp to datetime
#     df["open_time"] = pd.to_datetime(df["open_time"], unit="ms")

#     # Convert numeric columns to float
#     numeric_cols = ["open", "high", "low", "close"]
#     df[numeric_cols] = df[numeric_cols].astype(float)

#     # --- Calculate Bollinger Bands ---
#     window, std_dev = 20, 2
#     df["SMA"] = df["close"].rolling(window=window).mean()
#     df["Upper_BB"] = df["SMA"] + (df["close"].rolling(window=window).std() * std_dev)
#     df["Lower_BB"] = df["SMA"] - (df["close"].rolling(window=window).std() * std_dev)

#     # --- Calculate RSI ---
#     period = 14
#     delta = df["close"].diff(1)
#     gain = np.where(delta > 0, delta, 0)
#     loss = np.where(delta < 0, -delta, 0)

#     avg_gain = pd.Series(gain).ewm(span=period, adjust=False).mean()
#     avg_loss = pd.Series(loss).ewm(span=period, adjust=False).mean()

#     rs = avg_gain / avg_loss
#     df["RSI"] = 100 - (100 / (1 + rs))

#     # --- Identify Trade Signals ---
#     df["Long_Entry"] = (df["RSI"] <= 30) & (df["close"] < df["Lower_BB"])
#     df["Short_Entry"] = (df["RSI"] > 70) & (df["close"] > df["Upper_BB"])
#     df["Long_Exit"] = df["close"] >= df["Upper_BB"]
#     df["Short_Exit"] = df["close"] <= df["Lower_BB"]

#     # Convert timestamps to seconds
#     df["time"] = df["open_time"].astype("int64") // 10**9

#     # Remove duplicates and sort
#     df = df.drop_duplicates(subset=["time"]).sort_values(by="time", ascending=True)
#     df["Upper_BB"] = df["Upper_BB"].ffill()
#     df["Lower_BB"] = df["Lower_BB"].ffill()
#     df["RSI"] = df["RSI"].ffill()

#     # Count winning and losing trades
#     winning_trades = ((df["Long_Entry"] & df["Long_Exit"]) | (df["Short_Entry"] & df["Short_Exit"])).sum()
#     losing_trades = ((df["Long_Entry"] & ~df["Long_Exit"]) | (df["Short_Entry"] & ~df["Short_Exit"])).sum()

#     print(f"Winning trades: {winning_trades}")
#     print(f"Losing trades: {losing_trades}")

#     return df

# # Load Data
# eurusd = pd.read_excel("../database/eurusd/eurusd2022.xlsx")

# # Run Strategy
# bollingerbands = bollinger_bands_strategy(eurusd)
import pandas as pd
import numpy as np

def bollinger_bands_strategy(df: pd.DataFrame) -> pd.DataFrame:
    df.columns = ["open_time", "open", "high", "low", "close", "volume"]
    df = df.drop(columns=["volume"])  # Drop volume column

    # Convert timestamp to datetime
    df["open_time"] = pd.to_datetime(df["open_time"], unit="ms")

    # Convert numeric columns to float
    numeric_cols = ["open", "high", "low", "close"]
    df[numeric_cols] = df[numeric_cols].astype(float)

    # --- Calculate Bollinger Bands ---
    window, std_dev = 20, 2
    df["SMA"] = df["close"].rolling(window=window).mean()
    df["Upper_BB"] = df["SMA"] + (df["close"].rolling(window=window).std() * std_dev)
    df["Lower_BB"] = df["SMA"] - (df["close"].rolling(window=window).std() * std_dev)

    # --- Calculate RSI ---
    period = 14
    delta = df["close"].diff(1)
    gain = np.where(delta > 0, delta, 0)
    loss = np.where(delta < 0, -delta, 0)

    avg_gain = pd.Series(gain).ewm(span=period, adjust=False).mean()
    avg_loss = pd.Series(loss).ewm(span=period, adjust=False).mean()

    rs = avg_gain / avg_loss
    df["RSI"] = 100 - (100 / (1 + rs))

    # --- Identify Trade Signals ---
    df["Long_Entry"] = (df["RSI"] <= 30) & (df["close"] < df["Lower_BB"])
    df["Short_Entry"] = (df["RSI"] > 70) & (df["close"] > df["Upper_BB"])
    df["Long_Exit"] = df["close"] >= df["Upper_BB"]
    df["Short_Exit"] = df["close"] <= df["Lower_BB"]

    # Convert timestamps to seconds
    df["time"] = df["open_time"].astype("int64") // 10**9

    # Remove duplicates and sort
    df = df.drop_duplicates(subset=["time"]).sort_values(by="time", ascending=True)
    df["Upper_BB"] = df["Upper_BB"].ffill()
    df["Lower_BB"] = df["Lower_BB"].ffill()
    df["RSI"] = df["RSI"].ffill()

    # Count winning and losing trades
    winning_trades = ((df["Long_Entry"] & df["close"] >= df["Upper_BB"]) | (df["Short_Entry"] & df["close"] <= df["Lower_BB"])).sum()
    losing_trades = ((df["Long_Entry"] & df["close"] < df["Upper_BB"]) | (df["Short_Entry"] & df["close"] > df["Lower_BB"])).sum()

    print(f"Winning trades: {winning_trades}")
    print(f"Losing trades: {losing_trades}")

    return df

# Load Data
eurusd = pd.read_excel("../database/eurusd/eurusd2022.xlsx")

# Run Strategy
bollingerbands = bollinger_bands_strategy(eurusd)

