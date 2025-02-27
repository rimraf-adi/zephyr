from dataingestion import BinanceDataIngestor
from enums import Interval
from config import API_KEY, API_SECRET

interval = Interval.D1
client = BinanceDataIngestor(API_KEY, API_SECRET)
df = client.get_data(interval, "BTCUSDT", "2024-01-01", "2024-12-31")
df.to_csv(f"../database/BTCUSDT/{interval.value}.csv")
# print(df)