import re
import matplotlib.pyplot as plt
import numpy as np
import sys


def convert_to_nanoseconds(time_string):
    # Define conversion factors
    unit_factors = {'ns': 1, 'µs': 1000, 'ms': 1000000}

    # Extract value and unit using regular expression
    match = re.search(r'(\d+[\.\d]*)([mµn]s)', time_string)
    
    if match:
        value, unit = match.groups()
        value = float(value)
        
        # If unit is 'ns', just return the value
        if unit == 'ns':
            return value
        # If unit is 'ms', convert to nanoseconds
        elif unit == 'ms':
            return value * unit_factors['ms']
        # If unit is 'µs', convert to nanoseconds
        elif unit == 'µs':
            return value * unit_factors['µs']
        else:
            raise ValueError("Invalid time unit. Supported units: ns, µs, ms")
    else:
        raise ValueError("Invalid time string format")


def process_text_file(file_path, target_string):
    matching_lines = []

    with open(file_path, 'r') as file:
        for line_number, line in enumerate(file, start=1):
            if target_string in line:
                match = re.search(r'\t.*', line)
                str = match.string
                split = re.split(r'\t', str)
                value = convert_to_nanoseconds(split[1].replace("\n", ""))
                matching_lines.append((line_number, line.strip(), float(value)))

    return matching_lines

def plot_graph(data):
    line_numbers, matching_lines, extracted_values = zip(*data)


    z = np.polyfit(line_numbers, extracted_values, 1)
    p = np.poly1d(z)

    plt.plot(line_numbers, extracted_values, marker='o', linestyle='-', color='b')
    plt.xlabel('Line Number')
    plt.ylabel('Extracted Values')
    plt.title('Match over iterations')

    plt.plot(line_numbers, p(line_numbers), "r--")

    plt.savefig("output.png")

if __name__ == "__main__":
    # file_path = input("Enter the path to the text file: ")
    # target_string = input("Enter the target string to match: ")

    # file_path =
    # target_string = sys.argv[1]

    file_path = "/home/esiramos/projects/grpchan/shmgrpc/time.txt"
    target_string = "find unary"


    matching_lines = process_text_file(file_path, target_string)

    if matching_lines:
        print("Matching lines and extracted values:")
        for line_number, line_content, extracted_value in matching_lines:
            print(f"Line {line_number}: {line_content} | Extracted Value: {extracted_value}")

        plot_graph(matching_lines)
    else:
        print(f"No matching lines found for '{target_string}'.")
