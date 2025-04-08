# get abs path of this file
import argparse
import os
from pathlib import Path

WORKDIR = Path(__file__).parent.resolve()

def main():
    # if arg is provided, should be a *.hl7 file
    # default to iterate over all *.hl7 files in current directory
    parser = argparse.ArgumentParser(description='Process some HL7 files.')
    parser.add_argument('file', type=str, nargs='?', default=None,)
    
    args = parser.parse_args()

    if args.file:
        filepath = WORKDIR / args.file
        print(f"Processing file: {filepath}")
        try:
            clean_file(filepath)
        except FileNotFoundError:
            print(f"File {filepath} not found.")
            return
    
    for filename in os.listdir(WORKDIR):
        if not filename.endswith('.hl7'):
            continue

        filepath = WORKDIR / filename
        try:
            clean_file(filepath)
        except FileNotFoundError:
            print(f"File {filepath} not found.")
            continue

def clean_file(filepath: str) -> None:
    with open(filepath, 'rb') as f:
        hl7 = f.read()
            
    fixed = hl7.replace(b'\r\n', b'\r')
    with open(filepath, 'wb') as f:
        f.write(fixed)

main()
