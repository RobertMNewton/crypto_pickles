"""
Example usage of the orderbook_reader module
"""

from orderbook_reader import load_dataset, load_orderbook_file
from datetime import datetime


def example_single_file():
    """Example: Load and process a single file"""
    print("=== Single File Example ===\n")
    
    # Load a single file
    hist = load_orderbook_file("data/json/btcusdt/1707318000000-1707318300000.json")
    
    print(f"Symbol: {hist.symbol}")
    print(f"Time range: {hist.get_time_range()}")
    print(f"Number of frames: {len(hist.timestamps)}")
    
    # Get orderbook at start
    ob = hist.get_orderbook_at_index(0)
    print(f"\nOrderbook at start:")
    print(f"  Time: {datetime.fromtimestamp(ob.time / 1000)}")
    print(f"  Best bid: {max(ob.bids.keys()):.2f}")
    print(f"  Best ask: {min(ob.asks.keys()):.2f}")
    print(f"  Spread: {ob.get_spread():.2f}")
    print(f"  Mid price: {ob.get_mid_price():.2f}")
    
    # Get top 5 levels
    print(f"\n  Top 5 bids:")
    for price, volume in ob.get_sorted_bids(5):
        print(f"    {price:.2f} @ {volume:.6f}")
    
    print(f"\n  Top 5 asks:")
    for price, volume in ob.get_sorted_asks(5):
        print(f"    {price:.2f} @ {volume:.6f}")


def example_dataset():
    """Example: Query a dataset directory"""
    print("\n\n=== Dataset Example ===\n")
    
    # Load entire dataset
    dataset = load_dataset("data/json", symbol="btcusdt")
    
    print(f"Found {len(dataset.files)} files")
    
    # Get available range
    start, end = dataset.get_available_range()
    print(f"\nData available from {start} to {end}")
    
    # List all files
    print("\nFiles in dataset:")
    for file_info in dataset.list_files()[:5]:  # Show first 5
        print(f"  {file_info['start']} to {file_info['end']}")
    
    # Get orderbook at specific time
    if dataset.file_ranges:
        first_timestamp = dataset.file_ranges[0][0]
        ob = dataset.get_orderbook_at_time(first_timestamp, depth=10)
        
        if ob:
            print(f"\nOrderbook at {ob['timestamp']}:")
            print(f"  Spread: {len(ob['bids'])} bids, {len(ob['asks'])} asks")
            print(f"  Best bid: {ob['bids'][0][0]:.2f}")
            print(f"  Best ask: {ob['asks'][0][0]:.2f}")


def example_time_range_query():
    """Example: Query orderbooks in a time range"""
    print("\n\n=== Time Range Query Example ===\n")
    
    dataset = load_dataset("data/json", symbol="btcusdt")
    
    if not dataset.file_ranges:
        print("No data available")
        return
    
    # Get first 5 minutes of data
    start_time = dataset.file_ranges[0][0]
    end_time = start_time + (5 * 60 * 1000)  # 5 minutes in milliseconds
    
    # Query with sampling every 10 frames, depth of 5 levels
    orderbooks = dataset.get_orderbooks_in_range(
        start_time, 
        end_time, 
        freq=10,  # Sample every 10th frame
        depth=5   # Top 5 levels only
    )
    
    print(f"Retrieved {len(orderbooks)} orderbooks")
    
    if orderbooks:
        print(f"\nFirst orderbook:")
        print(f"  Time: {orderbooks[0]['timestamp']}")
        print(f"  Best bid: {orderbooks[0]['bids'][0]}")
        print(f"  Best ask: {orderbooks[0]['asks'][0]}")
        
        print(f"\nLast orderbook:")
        print(f"  Time: {orderbooks[-1]['timestamp']}")
        print(f"  Best bid: {orderbooks[-1]['bids'][0]}")
        print(f"  Best ask: {orderbooks[-1]['asks'][0]}")


def example_reconstruct_all():
    """Example: Reconstruct all orderbooks from a file"""
    print("\n\n=== Reconstruct All Example ===\n")
    
    # Load a file
    hist = load_orderbook_file("data/json/btcusdt/1707318000000-1707318300000.json")
    
    # Get all orderbooks
    all_orderbooks = hist.get_all_orderbooks()
    
    print(f"Reconstructed {len(all_orderbooks)} orderbooks")
    
    # Calculate some statistics
    spreads = [ob.get_spread() for ob in all_orderbooks]
    mid_prices = [ob.get_mid_price() for ob in all_orderbooks]
    
    print(f"\nSpread statistics:")
    print(f"  Min: {min(spreads):.2f}")
    print(f"  Max: {max(spreads):.2f}")
    print(f"  Avg: {sum(spreads)/len(spreads):.2f}")
    
    print(f"\nPrice movement:")
    print(f"  Start: {mid_prices[0]:.2f}")
    print(f"  End: {mid_prices[-1]:.2f}")
    print(f"  Change: {mid_prices[-1] - mid_prices[0]:.2f}")


if __name__ == "__main__":
    # Run examples (comment out ones you don't need)
    
    try:
        # example_single_file()
        example_dataset()
        # example_time_range_query()
        # example_reconstruct_all()
    except FileNotFoundError as e:
        print(f"Error: {e}")
        print("\nMake sure you have collected some data first:")
        print("  go run cmd/dataminer/main.go -config=config/miner/local-json.yaml")
