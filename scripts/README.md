# Orderbook Data Scripts

Python utilities for reading and processing CryptoPickles orderbook data.

## Installation

No external dependencies required - uses only Python standard library.

## Quick Start

### Load a dataset directory

```python
from scripts.orderbook_reader import load_dataset

# Load all files for a symbol
dataset = load_dataset("data/json", symbol="btcusdt")

# See what's available
print(dataset.get_available_range())
print(dataset.list_files())

# Get orderbook at specific time (milliseconds)
ob = dataset.get_orderbook_at_time(1707318000000, depth=10)
print(ob['bids'][:5])  # Top 5 bid levels
print(ob['asks'][:5])  # Top 5 ask levels

# Get orderbooks in a time range
orderbooks = dataset.get_orderbooks_in_range(
    start_time=1707318000000,
    end_time=1707318300000,
    freq=10,   # Sample every 10th frame
    depth=20   # Top 20 levels
)
```

### Load a single file

```python
from scripts.orderbook_reader import load_orderbook_file

hist = load_orderbook_file("data/json/btcusdt/1707318000000-1707318300000.json")

# Get orderbook at specific index
ob = hist.get_orderbook_at_index(0)  # Start
ob = hist.get_orderbook_at_index(100)  # 100th frame

# Get orderbook at or before timestamp
ob = hist.get_orderbook_at_time(1707318150000)

# Reconstruct all orderbooks
all_obs = hist.get_all_orderbooks()
```

## API Reference

### OrderBookDataset

Main class for working with a directory of orderbook files.

**Methods:**
- `get_orderbook_at_time(timestamp, depth=None)` - Get orderbook at specific time
- `get_orderbooks_in_range(start, end, freq=1, depth=None)` - Get multiple orderbooks
- `get_available_range()` - Get (start_datetime, end_datetime) tuple
- `list_files()` - List all files with time ranges
- `get_file_for_time(timestamp)` - Find file containing timestamp

### OrderBookHistory

Class for working with a single orderbook file.

**Methods:**
- `get_orderbook_at_index(index)` - Get orderbook at frame index
- `get_orderbook_at_time(timestamp)` - Get orderbook at or before timestamp
- `get_all_orderbooks()` - Reconstruct all orderbooks in file
- `get_time_range()` - Get (start_ms, end_ms) tuple

### OrderBook

Class representing a full orderbook state.

**Methods:**
- `get_sorted_bids(depth=None)` - Get bids as [(price, volume), ...]
- `get_sorted_asks(depth=None)` - Get asks as [(price, volume), ...]
- `to_dict(depth=None)` - Convert to dictionary
- `get_spread()` - Get bid-ask spread
- `get_mid_price()` - Get mid price

**Attributes:**
- `time` - Timestamp in milliseconds
- `bids` - Dict of {price: volume}
- `asks` - Dict of {price: volume}

## Examples

See [example_usage.py](example_usage.py) for complete examples.

### Calculate spreads over time

```python
hist = load_orderbook_file("data/json/btcusdt/1707318000000-1707318300000.json")
all_obs = hist.get_all_orderbooks()

spreads = [ob.get_spread() for ob in all_obs]
times = [ob.time for ob in all_obs]

# Now plot with matplotlib, analyze, etc.
```

### Export to CSV

```python
import csv

dataset = load_dataset("data/json", symbol="btcusdt")
orderbooks = dataset.get_orderbooks_in_range(start, end, depth=10)

with open('orderbooks.csv', 'w') as f:
    writer = csv.writer(f)
    writer.writerow(['timestamp', 'best_bid', 'best_ask', 'spread'])
    
    for ob in orderbooks:
        best_bid = ob['bids'][0][0] if ob['bids'] else 0
        best_ask = ob['asks'][0][0] if ob['asks'] else 0
        writer.writerow([ob['time'], best_bid, best_ask, best_ask - best_bid])
```

### Find price at specific time

```python
dataset = load_dataset("data/json", symbol="btcusdt")

# Convert datetime to milliseconds
from datetime import datetime
dt = datetime(2026, 2, 7, 12, 30, 0)
timestamp_ms = int(dt.timestamp() * 1000)

ob = dataset.get_orderbook_at_time(timestamp_ms)
if ob:
    print(f"Mid price at {dt}: {(ob['bids'][0][0] + ob['asks'][0][0]) / 2}")
```

## Command Line Usage

```bash
# List files in dataset
python scripts/orderbook_reader.py data/json btcusdt

# Run examples
python scripts/example_usage.py
```
