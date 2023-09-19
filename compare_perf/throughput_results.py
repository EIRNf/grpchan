import subprocess

# Define the server and client commands
# server_command = "go run /home/estebanramos/app_testing/grpchan/shmgrpc/test_server/main.go "
# client_command = "go test /home/estebanramos/app_testing/grpchan/shmgrpc/test_client/client_bench_test.go  -bench=."

server_command = "go run /home/estebanramos/app_testing/grpchan/http2/server/main.go "
client_command = "go test /home/estebanramos/app_testing/grpchan/http2/client/client_test.go -bench=."

# Number of iterations
num_iterations = 10

# Define arrays to store the outputs for each run
server_outputs = []
client_outputs = []
client_outputs1 = []
client_outputs2 = []



# Define the strings to match in the output
# string_to_match_server = "Server output: "
string_to_match_client = "Throughput: "
string_to_match_client1 = "Mean: "
string_to_match_client2 = "Time in Seconds: "



def run_command_no_output(command):
    try:
        result = subprocess.Popen(command, stdout=subprocess.PIPE, stderr=subprocess.PIPE, shell=True, text=True)
        return result.stdout
    except subprocess.CalledProcessError as e:
        print(f"Error running command: {e}")
        return ""


# Function to run a command and capture its output
def run_command(command):
    try:
        result = subprocess.run(command, stdout=subprocess.PIPE, stderr=subprocess.PIPE, shell=True, text=True)
        return result.stdout
    except subprocess.CalledProcessError as e:
        print(f"Error running command: {e}")
        return ""

def Average(lst):
    return sum(lst) / len(lst)

# Run the server and client binaries in a loop
for _ in range(num_iterations):
    # Start the server
    server_output = run_command_no_output(server_command)

    # Start the client
    client_output = run_command(client_command)

    client_result  = ""
    client_result1  = ""
    client_result2  = ""

    for line in client_output.splitlines():
        if string_to_match_client in line:
            client_result += line.replace(string_to_match_client, "") + "\n"
        if string_to_match_client1 in line:
            client_result1 += line.replace(string_to_match_client1, "") + "\n"
        if string_to_match_client2 in line:
            client_result2 += line.replace(string_to_match_client2, "") + "\n"


    # Append the outputs to the arrays
    # server_outputs.append(server_output)
    client_outputs.append(float(client_result))
    client_outputs1.append(float(client_result1))
    client_outputs2.append(float(client_result2))


# Print the aggregated results for each run
for i in range(num_iterations):
    # print(f"Run {i + 1} - Server Results:")
    # print(server_outputs[i])

    print(f"\nRun {i + 1} - Client Results:")
    print(client_outputs[i])
    print(client_outputs1[i])
    print(client_outputs2[i])


#Print Averaged results

F("\nAverage for {num_iterations} Runs - Client Results:")
print(Average(client_outputs))
print(Average(client_outputs1))
print(Average(client_outputs2))
