// main
package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"math"
	"math/rand"
	"os"
	"sync"
	"time"
	"strconv"
)

// Wait group for our thread pool.
var wg sync.WaitGroup

var debug bool = true

/*
Random Helper Functions
*/

// Find the minimum of an int array.
func min(s []int) (int, int) {

	var minimum int = 100000
	var mindex int = 0
	for i := 0; i < len(s); i++ {
		if s[i] < minimum {
			minimum = s[i]
			mindex = i
		}
	}

	return minimum, mindex
}

// Removes element from int array. Taken from: https://yourbasic.org/golang/delete-element-slice/
func remove(a []int, i int) []int {
	copy(a[i:], a[i+1:]) // Shift a[i+1:] left one index.
	a[len(a)-1] = 0      // Erase last element (write zero value).
	a = a[:len(a)-1]     // Truncate slice.

	return a
}

/*
Message struct + methods.

Implements the first and source/dest updating methods from the algo.
*/
type message struct {
	mtype    string
	value    int
	stage    int
	source   []bool
	dest     []bool
	prevLink int
	hops     int
}

// This method adds a source path when we traverse a link.
func (m *message) update_source(link int) {
	if m.source[link] == true {
		m.source[link] = false
	} else {
		m.source[link] = true
	}
}

// This method pops the next destination
func (m *message) first() int {

	var out int = -1

	for i := 0; i < len(m.dest); i++ {

		if m.dest[i] == true {
			out = i
			m.dest[i] = false
			break
		}
	}

	return out
}

// This checks to see if the destination is empty.
func (m *message) check_dest() bool {

	var out bool = false

	for i := 0; i < len(m.dest); i++ {
		if m.dest[i] == true {
			out = true
			break
		}
	}

	return out
}

/*
Node struct + methods
*/

type node struct {
	id              int
	state           string
	stage           int
	delay           []int
	input           chan *message
	Delayed         []*message
	links           []chan *message
	NextDuelist     []bool
	k               int
	n_transmissions int
	n_hyper         int
	leader          int
}

// Constructor. Decrements k/stage by one for stage/link naming purposes to play nicely with 0-indexed arrays.
func NewNode(_id int, _k int) *node {

	new_node := &node{id: _id, state: "ASLEEP", stage: 0, delay: make([]int, 0), input: make(chan *message, _k*(_k+2)), Delayed: make([]*message, 0), links: make([]chan *message, 0), NextDuelist: make([]bool, _k), k: _k - 1, n_transmissions: 0, n_hyper: 0, leader: -1}

	return new_node
}

// Print node contents for diagnostic purposes.
func (nd *node) PRINT(bin int) {

	width := nd.k + 1

	fmt.Printf("Coordinate: %0*b  ID: %d  State: %s  Stage:  %d  Leader: %d  Messages Transmitted: %d\n", width, bin, nd.id, nd.state, nd.stage+1, nd.leader, nd.n_transmissions)
}

// helper to get message from Delayed slice.
func (nd *node) popMessage(stage int) *message {

	var returnMessage *message = &message{}
	var remIndex int = 0

	if stage == -1 {
		returnMessage = nd.Delayed[0]
	} else {
		for i := 0; i < len(nd.Delayed); i++ {
			if nd.Delayed[i].stage == stage {
				returnMessage = nd.Delayed[i]
				remIndex = i
				break
			}
		}
	}

	copy(nd.Delayed[remIndex:], nd.Delayed[remIndex+1:]) // Shift a[i+1:] left one index.
	nd.Delayed[len(nd.Delayed)-1] = &message{}           // Erase last element (write zero value).
	nd.Delayed = nd.Delayed[:len(nd.Delayed)-1]          // Truncate slice.

	return returnMessage
}

// Implements Hyperflood. Set link to -1 to notify all.
func (nd *node) NOTIFY(link int, leader int) {
	for i := link + 1; i < len(nd.links); i++ {
		nd.links[i] <- &message{mtype: "Notify", value: leader, stage: nd.k, source: nil, dest: nil, prevLink: i, hops: 0}
		nd.n_transmissions++
		nd.n_hyper++
	}
}

// Update minimum delay and add delayed message to delay list.
func (nd *node) DELAY_MESSAGE(m *message) {

	if debug {
		fmt.Printf("Node %d in Stage %d delayed a message from Node %d for Stage %d\n", nd.id, nd.stage+1, m.value, m.stage+1)
	}

	nd.delay = append(nd.delay, m.stage)

	nd.Delayed = append(nd.Delayed, m)
}

// Update minimum delay and add delayed message to delay list.
func (nd *node) PROCESS_MESSAGE(m *message) {

	if m.value > nd.id {
		if m.stage == nd.k {
			nd.leader = nd.id
			nd.NOTIFY(-1, nd.leader)
			nd.state = "LEADER"

			if debug {
				fmt.Printf("Node %d won (%d hops).\n", nd.id, m.hops)
			}

		} else {
			nd.stage = nd.stage + 1

			if debug {
				fmt.Printf("Node %d reached Stage %d by defeating Node %d in Stage %d (%d hops). Sending duel request across Link %d \n", nd.id, nd.stage+1, m.value, m.stage+1, m.hops, nd.stage+1)
			}

			new_message := &message{mtype: "Match", value: nd.id, stage: nd.stage, source: make([]bool, nd.k+1), dest: make([]bool, nd.k+1), prevLink: nd.stage, hops: 1}
			new_message.update_source(nd.stage)

			nd.links[nd.stage] <- new_message
			nd.n_transmissions++
			nd.CHECK()
		}
	} else {

		copy(nd.NextDuelist, m.source)

		if debug {
			fmt.Printf("Node %d was defeated by Node %d in Stage %d (%d hops). Next Duelist at v+%\n", nd.id, m.value, m.stage+1, m.hops, m.source)
		}

		nd.CHECK_ALL()
		nd.state = "DEFEATED"
	}

}

func (nd *node) CHECK() {

	if len(nd.Delayed) > 0 {
		minimum, mindex := min(nd.delay)

		if minimum == nd.stage {
			nd.delay = remove(nd.delay, mindex)
			nd.PROCESS_MESSAGE(nd.popMessage(minimum))
		}
	}
}

func (nd *node) CHECK_ALL() {

	for len(nd.Delayed) != 0 {
		newMessage := nd.popMessage(-1)
		if !newMessage.check_dest() {
			copy(newMessage.dest, nd.NextDuelist)
		}

		new_link := newMessage.first()
		newMessage.update_source(new_link)
		newMessage.hops += 1
		nd.links[new_link] <- newMessage

		if debug {
			fmt.Printf("Node %d routed a deferred duel request from Node %d at Stage %d across Link %d\n", nd.id, newMessage.value, newMessage.stage+1, new_link+1)
		}

		nd.n_transmissions++
	}
}

// Spins up a node and runs it to completion. Designed to be used in goroutine.
func (nd *node) launch(w *sync.WaitGroup) {

	if debug {
		fmt.Printf("Starting Node %d\n", nd.id)
	}

	defer w.Done()

	for nd.state != "FOLLOWER" && nd.state != "LEADER" {

		if nd.state == "ASLEEP" {
			new_message := &message{mtype: "Match", value: nd.id, stage: nd.stage, source: make([]bool, nd.k+1), dest: make([]bool, nd.k+1), prevLink: nd.stage, hops: 1}
			new_message.update_source(nd.stage)
			nd.links[0] <- new_message
			nd.state = "DUELIST"
			nd.n_transmissions++

			if debug {
				fmt.Printf("Node %d sent an initial duel request across Link %d\n", nd.id, 1)
			}
		} else if nd.state == "DUELIST" {

			received_message := <-nd.input

			if received_message.mtype == "Match" {
				if received_message.stage == nd.stage {
					nd.PROCESS_MESSAGE(received_message)
				} else {
					nd.DELAY_MESSAGE(received_message)
				}
			} else if received_message.mtype == "Notify" {

				if debug {
					fmt.Printf("Node %d got the Notify message before finishing its duel and will now forward it.\n", nd.id)
				}

				nd.leader = received_message.value
				nd.NOTIFY(received_message.prevLink, received_message.value)
				nd.state = "FOLLOWER"
			}

		} else if nd.state == "DEFEATED" {
			received_message := <-nd.input

			if received_message.mtype == "Match" {
				if !received_message.check_dest() {

					if debug {
						fmt.Printf("Node %d received a duel request from Node %d at Stage %d with an empty destination and routed it to %v\n", nd.id, received_message.value, received_message.stage+1, nd.NextDuelist)
					}

					copy(received_message.dest, nd.NextDuelist)
				}

				new_link := received_message.first()
				received_message.update_source(new_link)
				received_message.hops += 1
				nd.links[new_link] <- received_message
				nd.n_transmissions++

				if debug {
					fmt.Printf("Defeated Node %d routed a duel request from Node %d for Stage %d across Link %d\n", nd.id, received_message.value, received_message.stage+1, new_link+1)
				}
			}

			if received_message.mtype == "Notify" {
				nd.leader = received_message.value
				nd.NOTIFY(received_message.prevLink, received_message.value)
				nd.state = "FOLLOWER"
			}
		}
	}

	if debug {
		fmt.Printf("Node %d finished\n", nd.id)
	}
}

/*
This generates a hypercube of size 2^n and runs the algorithm.
ids are randomly generated and assigned.
returns upper bound (formula from textbook), experimental message
complexity, and time taken.
*/
func run_experiment(k int) (int64, int64, int64, int64, int64) {

	n := math.Pow(2.0, float64(k))
	wg.Add(int(n))

	id_array := make([]int, uint(n))

	// Generate id array
	for i := 0; i < len(id_array); i++ {
		id_array[i] = i
	}

	// Shuffle array randomly, seeded by current time.
	// We sleep for 5 ms to avoid throwing off the RNG.
	time.Sleep(5 * time.Millisecond)
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(id_array), func(i, j int) { id_array[i], id_array[j] = id_array[j], id_array[i] })

	// Build array of Hypercube nodes.
	arr := make([]*node, uint(n))

	for i := 0; i < len(arr); i++ {

		arr[i] = NewNode(id_array[i], k)
	}

	/*
		Builds linkage. Uses Gray Code coordinates to determine neighbours.

		Done by using shifts to get 0001, 0010, etc. and XOR with current index
		to get the target for each link.
	*/

	for i := 0; i < len(arr); i++ {

		_links := make([]chan *message, k)

		var shift_amount int = 1

		for j := 0; j < k; j++ {
			target_ind := i ^ shift_amount
			_links[j] = arr[target_ind].input

			shift_amount = shift_amount << 1

		}

		arr[i].links = _links
	}

	// Diagnostics to see what happened with all the nodes.
	if debug {
		fmt.Println("Node Configuration:")
		for i := 0; i < len(arr); i++ {
			arr[i].PRINT(i)
		}

		fmt.Println("\n")
	}

	if debug {
		fmt.Println("Launching threads...")
	}

	start := time.Now().UnixNano()
	// Launches all threads
	for i := 0; i < len(arr); i++ {
		go arr[i].launch(&wg)
	}

	// Waits for algorithm to finish.
	wg.Wait()

	finish := time.Now().UnixNano()
	resultant_time := finish - start

	// Compute formula for upper bound
	messages_form := int(7*n - math.Pow(math.Log2(n), 2.0) - 3*math.Log2(n) - 7)
	messages_exp := 0
	messages_hyper := 0

	var success int64 = 1
	for i := 0; i < len(arr); i++ {
		messages_exp += arr[i].n_transmissions
		messages_hyper += arr[i].n_hyper

		if arr[i].leader != 0 {
			success = 0
		}
	}

	// Diagnostics to see what happened with all the nodes.
	if debug {
		for i := 0; i < len(arr); i++ {
			arr[i].PRINT(i)
		}
	}

	return int64(messages_form), int64(messages_exp), int64(messages_hyper), resultant_time, success
}

func main() {

	k := flag.Int("k", 5, "The dimension of the hypercube we wish to experiment on.")
	k_flag := flag.Bool("uptok", false, "Controls if we sample from 2 up to k, as opposed to just k.")
	samples := flag.Int("samples", 100, "The number of experiments to run for each k.")
	debug_flag := flag.Bool("debug", false, "Controls if we print debug statements that trace the execution of the algorithm. Will get very verbose for high k.")

	flag.Parse()

	debug = *debug_flag

	k_begin := *k
	uptok_string := ""

	if *k_flag {
		k_begin = 2
		uptok_string = "upto"
	}

	// Create output file,with suffix based on current system time.
	fileName := fmt.Sprintf("results_%sk%s_%s.csv", uptok_string, strconv.Itoa(*k), strconv.FormatInt(time.Now().Unix(), 10))

	csvFile, err := os.Create(fileName)

	if err != nil {
		fmt.Printf("failed creating file: %s\n", err)
		return
	}

	fmt.Println("Beginning Simulation...")

	// The CSV header.
	cols := []string{"k", "Messages", "UpperBound", "Time", "Success"}

	// Prep file writer
	csvwriter := csv.NewWriter(csvFile)
	_ = csvwriter.Write(cols)

	for i := k_begin; i <= *k; i++ {
		for j := 0; j < *samples; j++ {

			var upper_bound, messages_sent, messages_hyper, duration, success = run_experiment(i)

			if debug {
				fmt.Printf("Messages (Total): %d, Upper Bound (Total): %d,  Messages (HyperFlood): %d, HyperFlood (Expected): %d  Time Taken: %v\n\n", messages_sent, upper_bound, messages_hyper, int(math.Pow(2.0, float64(*k)))-1, duration)
			}

			outRow := []string{strconv.FormatInt(int64(i), 10), strconv.FormatInt(messages_sent, 10), strconv.FormatInt(upper_bound, 10), strconv.FormatInt(duration, 10), strconv.FormatInt(success, 10)}
			_ = csvwriter.Write(outRow)

		}

		fmt.Printf("Generated %d samples for a hypercube of size %d.\n", *samples, i)
	}

	csvwriter.Flush()
	csvFile.Close()

	fmt.Printf("Wrote output file %s\n", fileName)
}
