import pandas as pd
import numpy as np

def trading_strategy(df: pd.DataFrame) -> pd.DataFrame:
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

    # --- Calculate MACD ---
    df["EMA_12"] = df["close"].ewm(span=12, adjust=False).mean()
    df["EMA_26"] = df["close"].ewm(span=26, adjust=False).mean()
    df["MACD"] = df["EMA_12"] - df["EMA_26"]
    df["Signal_Line"] = df["MACD"].ewm(span=9, adjust=False).mean()

    # --- Calculate Rate of Change (Momentum) ---
    df["ROC"] = df["close"].pct_change(periods=10) * 100

    # --- Calculate ATR (Volatility) ---
    df["TR"] = np.maximum(df["high"] - df["low"], 
                           np.maximum(abs(df["high"] - df["close"].shift(1)),
                                      abs(df["low"] - df["close"].shift(1))))
    df["ATR"] = df["TR"].rolling(window=14).mean()

    # --- Identify Trade Signals ---
    df["Long_Entry"] = (df["MACD"] > df["Signal_Line"]) & (df["RSI"] <= 30) & (df["ROC"] > 0)
    df["Short_Entry"] = (df["MACD"] < df["Signal_Line"]) & (df["RSI"] >= 70) & (df["ROC"] < 0)
    
    df["Long_Exit"] = (df["MACD"] < df["Signal_Line"]) | (df["close"] >= df["SMA"])
    df["Short_Exit"] = (df["MACD"] > df["Signal_Line"]) | (df["close"] <= df["SMA"])

    # Convert timestamps to seconds
    df["time"] = df["open_time"].astype("int64") // 10**9

    # Remove duplicates and sort
    df = df.drop_duplicates(subset=["time"]).sort_values(by="time", ascending=True)
    df[["Upper_BB", "Lower_BB", "RSI", "MACD", "Signal_Line", "ROC", "ATR"]] = df[["Upper_BB", "Lower_BB", "RSI", "MACD", "Signal_Line", "ROC", "ATR"]].ffill()

    # Count winning and losing trades
    winning_trades = ((df["Long_Entry"] & df["close"] >= df["SMA"]) | (df["Short_Entry"] & df["close"] <= df["SMA"])).sum()
    losing_trades = ((df["Long_Entry"] & df["close"] < df["SMA"]) | (df["Short_Entry"] & df["close"] > df["SMA"])).sum()
    
    total_trades = winning_trades + losing_trades
    accuracy = (winning_trades / total_trades) * 100 if total_trades > 0 else 0
    
    print(f"Winning trades: {winning_trades}")
    print(f"Losing trades: {losing_trades}")
    print(f"Accuracy: {accuracy:.2f}%")

    return df

# Load Data
eurusd = pd.read_excel("../database/eurusd/eurusd2022.xlsx")

# Run Strategy
result = trading_strategy(eurusd)
