# HyperElect-Simulation
A Go program that simulates oriented hypercube networks and runs the HyperElect leader election algorithm to determine message and time complexity relative to the algorithm's theoretical upper bound. Oriented hypercubes of size 2^k are created where individual nodes are represented as goroutines that communicate with each other across their respective links. Reference: https://doi.org/10.1006/jpdc.1996.0026
 
To run the program, navigate to the program directory and run main.go.exe. By default, the program 
runs 100 simulations for a hypercube of dimension 5 and does not print debug statements. However, optional 
command-line flags can be specified to change the program operaƟon, of the form -flag=value. These 
are as follows: 

-k: An integer specifying the highest value of k to run simulations for. Default: 5 

-uptok: A Boolean (true or false). If true, we run simulations for all k-values from 2 to k. If false, we just 
run simulations for the provided k-value. k-values of 0 and 1 are not considered as they’re trivial to 
compute. Default: false 

-samples: An integer specifying the number of simulaƟons to run for each k-value. Default: 100 

-debug: A Boolean (true or false). If true, we print debug statements tracing the execution of HyperElect, as well as initial and final configurations of all nodes. If false, we do not print debug 
statements. Default: false 

Note 1: Debug statements should not be printed for very large values of k or large numbers of 
simulations, as they will overwhelm the output.

Note 2: Depending on the speed and memory limitations of your system, large k-values may be lengthy to 
compute or crash the program. 

# Example 
main.go.exe -k=20 -uptok=true -samples=1000 
This statement will run 1000 simulations for each k-value from 2 to 20. 

# Output 
After the program finishes, the CSV file containing the results will be output to the same directory as the 
program. The file name indicates the chosen k value and the “uptok” parameter (if used) and has the 
Unix time in seconds at which the program was started appended to it to ensure uniqueness.

# Visualization
A Jupyter Notebook is included that generates line graphs showing the message complexity of all 
simulations for each k in comparison to their upper bound, as well as the average time taken for each k, 
factor increases for time, and the ratio of upper bound to highest seen message complexity for each k. 
To run it, launch the notebook, replace the specified file in the fileName variable with one of your 
choosing, and run all cells. This will create a new directory with the file name (minus the extension) and 
write all generated plots to it. The notebook requires Python, numpy, pandas, and matplotlib. 
