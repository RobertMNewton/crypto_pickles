"""
CryptoPickles Orderbook Reader
Reconstructs full orderbooks from the hybrid snapshot+diff format
"""

import json
import os
from pathlib import Path
from typing import Dict, List, Tuple, Optional
from datetime import datetime
import bisect


class OrderBook:
    """Represents a full orderbook at a point in time"""
    
    def __init__(self, time: int, bids: Dict[float, float], asks: Dict[float, float]):
        self.time = time
        self.bids = bids.copy()
        self.asks = asks.copy()
    
    def apply_diff(self, diff: Dict) -> 'OrderBook':
        """Apply a depth diff to this orderbook"""
        self.time = diff['Time']
        
        # Apply bid changes
        for price_str, volume_str in diff['Bids'].items():
            price = float(price_str)
            volume = float(volume_str)
            if volume == 0:
                self.bids.pop(price, None)
            else:
                self.bids[price] = volume
        
        # Apply ask changes
        for price_str, volume_str in diff['Asks'].items():
            price = float(price_str)
            volume = float(volume_str)
            if volume == 0:
                self.asks.pop(price, None)
            else:
                self.asks[price] = volume
        
        return self
    
    def get_sorted_bids(self, depth: Optional[int] = None) -> List[Tuple[float, float]]:
        """Get bids sorted by price (highest first)"""
        sorted_bids = sorted(self.bids.items(), key=lambda x: x[0], reverse=True)
        return sorted_bids[:depth] if depth else sorted_bids
    
    def get_sorted_asks(self, depth: Optional[int] = None) -> List[Tuple[float, float]]:
        """Get asks sorted by price (lowest first)"""
        sorted_asks = sorted(self.asks.items(), key=lambda x: x[0])
        return sorted_asks[:depth] if depth else sorted_asks
    
    def to_dict(self, depth: Optional[int] = None) -> Dict:
        """Convert to dictionary format"""
        return {
            'time': self.time,
            'timestamp': datetime.fromtimestamp(self.time / 1000).isoformat(),
            'bids': self.get_sorted_bids(depth),
            'asks': self.get_sorted_asks(depth)
        }
    
    def get_spread(self) -> float:
        """Get the bid-ask spread"""
        if not self.bids or not self.asks:
            return 0.0
        best_bid = max(self.bids.keys())
        best_ask = min(self.asks.keys())
        return best_ask - best_bid
    
    def get_mid_price(self) -> float:
        """Get the mid price"""
        if not self.bids or not self.asks:
            return 0.0
        best_bid = max(self.bids.keys())
        best_ask = min(self.asks.keys())
        return (best_bid + best_ask) / 2


class OrderBookHistory:
    """Represents a file containing orderbook history"""
    
    def __init__(self, filepath: str):
        self.filepath = filepath
        with open(filepath, 'r') as f:
            data = json.load(f)
        
        self.symbol = data['Symbol']
        
        # Parse start orderbook
        start_data = data['Start']
        self.start_time = start_data['Time']
        bids = {float(k): float(v) for k, v in start_data['Bids'].items()}
        asks = {float(k): float(v) for k, v in start_data['Asks'].items()}
        
        self.history = data['History']
        self.timestamps = [self.start_time] + [diff['Time'] for diff in self.history]
        
        # Cache the start orderbook
        self._start_orderbook = OrderBook(self.start_time, bids, asks)
    
    def get_orderbook_at_index(self, index: int) -> OrderBook:
        """Get orderbook at a specific index (0 = start, 1 = first diff, etc.)"""
        if index < 0 or index > len(self.history):
            raise IndexError(f"Index {index} out of range [0, {len(self.history)}]")
        
        # Start with the initial orderbook
        bids = {float(k): float(v) for k, v in self._start_orderbook.bids.items()}
        asks = {float(k): float(v) for k, v in self._start_orderbook.asks.items()}
        ob = OrderBook(self.start_time, bids, asks)
        
        # Apply diffs up to the index
        for i in range(index):
            ob.apply_diff(self.history[i])
        
        return ob
    
    def get_orderbook_at_time(self, timestamp: int) -> Optional[OrderBook]:
        """Get orderbook at or before a specific timestamp (milliseconds)"""
        # Find the latest timestamp <= requested time
        idx = bisect.bisect_right(self.timestamps, timestamp) - 1
        
        if idx < 0:
            return None
        
        return self.get_orderbook_at_index(idx)
    
    def get_all_orderbooks(self) -> List[OrderBook]:
        """Reconstruct all orderbooks in the history"""
        orderbooks = []
        
        # Start with initial orderbook
        bids = {float(k): float(v) for k, v in self._start_orderbook.bids.items()}
        asks = {float(k): float(v) for k, v in self._start_orderbook.asks.items()}
        ob = OrderBook(self.start_time, bids, asks)
        orderbooks.append(OrderBook(ob.time, ob.bids.copy(), ob.asks.copy()))
        
        # Apply each diff
        for diff in self.history:
            ob.apply_diff(diff)
            orderbooks.append(OrderBook(ob.time, ob.bids.copy(), ob.asks.copy()))
        
        return orderbooks
    
    def get_time_range(self) -> Tuple[int, int]:
        """Get the start and end timestamps"""
        return (self.timestamps[0], self.timestamps[-1])


class OrderBookDataset:
    """Manages multiple orderbook history files from a directory"""
    
    def __init__(self, directory: str, symbol: Optional[str] = None):
        self.directory = Path(directory)
        self.symbol = symbol
        self.files = self._discover_files()
        self.file_ranges = self._build_index()
    
    def _discover_files(self) -> List[Path]:
        """Find all JSON files in the directory"""
        if self.symbol:
            pattern = f"{self.symbol}/*.json"
        else:
            pattern = "**/*.json"
        
        files = sorted(self.directory.glob(pattern))
        return files
    
    def _build_index(self) -> List[Tuple[int, int, Path]]:
        """Build an index of (start_time, end_time, filepath)"""
        ranges = []
        for filepath in self.files:
            # Parse timestamps from filename: {start}-{end}.json
            filename = filepath.stem
            try:
                start_str, end_str = filename.split('-')
                start = int(start_str)
                end = int(end_str)
                ranges.append((start, end, filepath))
            except ValueError:
                print(f"Warning: Skipping file with unexpected name format: {filepath}")
        
        return sorted(ranges, key=lambda x: x[0])
    
    def get_file_for_time(self, timestamp: int) -> Optional[Path]:
        """Find the file containing a specific timestamp"""
        for start, end, filepath in self.file_ranges:
            if start <= timestamp <= end:
                return filepath
        return None
    
    def get_orderbook_at_time(self, timestamp: int, depth: Optional[int] = None) -> Optional[Dict]:
        """Get orderbook at a specific timestamp"""
        filepath = self.get_file_for_time(timestamp)
        if not filepath:
            return None
        
        hist = OrderBookHistory(str(filepath))
        ob = hist.get_orderbook_at_time(timestamp)
        
        if ob:
            return ob.to_dict(depth)
        return None
    
    def get_orderbooks_in_range(self, start_time: int, end_time: int, 
                                 freq: int = 1, depth: Optional[int] = None) -> List[Dict]:
        """
        Get orderbooks in a time range
        
        Args:
            start_time: Start timestamp (ms)
            end_time: End timestamp (ms)
            freq: Frequency - sample every N frames (1 = all frames)
            depth: Number of price levels to include (None = all)
        """
        orderbooks = []
        
        # Find all files that overlap with the range
        relevant_files = [
            filepath for start, end, filepath in self.file_ranges
            if not (end < start_time or start > end_time)
        ]
        
        for filepath in relevant_files:
            hist = OrderBookHistory(str(filepath))
            all_obs = hist.get_all_orderbooks()
            
            for i, ob in enumerate(all_obs):
                if start_time <= ob.time <= end_time and i % freq == 0:
                    orderbooks.append(ob.to_dict(depth))
        
        return sorted(orderbooks, key=lambda x: x['time'])
    
    def get_available_range(self) -> Optional[Tuple[datetime, datetime]]:
        """Get the full time range of available data"""
        if not self.file_ranges:
            return None
        
        start = self.file_ranges[0][0]
        end = self.file_ranges[-1][1]
        
        return (
            datetime.fromtimestamp(start / 1000),
            datetime.fromtimestamp(end / 1000)
        )
    
    def list_files(self) -> List[Dict]:
        """List all files with their time ranges"""
        return [
            {
                'file': str(filepath),
                'start': datetime.fromtimestamp(start / 1000).isoformat(),
                'end': datetime.fromtimestamp(end / 1000).isoformat(),
                'start_ms': start,
                'end_ms': end
            }
            for start, end, filepath in self.file_ranges
        ]


# Convenience functions
def load_orderbook_file(filepath: str) -> OrderBookHistory:
    """Load a single orderbook history file"""
    return OrderBookHistory(filepath)


def load_dataset(directory: str, symbol: Optional[str] = None) -> OrderBookDataset:
    """Load an entire dataset from a directory"""
    return OrderBookDataset(directory, symbol)


if __name__ == "__main__":
    # Example usage
    import sys
    
    if len(sys.argv) < 2:
        print("Usage:")
        print("  python orderbook_reader.py <directory> [symbol]")
        print("\nExample:")
        print("  python orderbook_reader.py data/json btcusdt")
        sys.exit(1)
    
    directory = sys.argv[1]
    symbol = sys.argv[2] if len(sys.argv) > 2 else None
    
    dataset = load_dataset(directory, symbol)
    
    print(f"Found {len(dataset.files)} files")
    print(f"\nAvailable range: {dataset.get_available_range()}")
    print(f"\nFiles:")
    for file_info in dataset.list_files():
        print(f"  {file_info['start']} to {file_info['end']}")
