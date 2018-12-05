package main

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

const ChunkSize = 8192

func TestPeerster_FileSearch(t *testing.T) {
	//noinspection GoUnhandledErrorResult
	defer func() {
		os.Remove("_SharedFiles/aaa.txt")
		os.Remove("_SharedFiles/.meta/aaa.txt")

		os.Remove("_SharedFiles/catapult inferior siege weapon")
		os.Remove("_SharedFiles/.meta/catapult inferior siege weapon")

		os.Remove("_SharedFiles/irrelevant_file_1")
		os.Remove("_SharedFiles/.meta/irrelevant_file_1")

		os.Remove("_SharedFiles/irrelevant_file_2")
		os.Remove("_SharedFiles/.meta/irrelevant_file_2")

		os.Remove("_SharedFiles/irrelevant_file_3")
		os.Remove("_SharedFiles/.meta/irrelevant_file_3")

		os.Remove("_SharedFiles/ms word")
		os.Remove("_SharedFiles/.meta/ms word")

		os.Remove("_Downloads/something something.avi")
		os.Remove("_Downloads/.meta/something something.avi")

		os.Remove("_Downloads/indeed catapults are inferior siege weapon.txt")
		os.Remove("_Downloads/.meta/indeed catapults are inferior siege weapon.txt")
	}()

	compile(t)

	// C (hits file1)
	// ^
	// B (hits file1) <- A   F (hits file2)
	// v                     ^
	// D----------E----------+

	aAddr := "localhost:2000"
	bAddr := "localhost:2001"
	cAddr := "localhost:2002"
	dAddr := "localhost:2003"
	eAddr := "localhost:2004"
	fAddr := "localhost:2005"

	a := createPeersterProcess(
		"8000",
		aAddr,
		"A",
		[]string{bAddr},
		"1",
		false,
	)
	b := createPeersterProcess(
		"8001",
		bAddr,
		"B",
		[]string{cAddr, dAddr},
		"1",
		false,
	)
	c := createPeersterProcess(
		"8002",
		cAddr,
		"C",
		[]string{},
		"1",
		false,
	)
	d := createPeersterProcess(
		"8003",
		dAddr,
		"D",
		[]string{eAddr},
		"1",
		false,
	)
	e := createPeersterProcess(
		"8004",
		eAddr,
		"E",
		[]string{fAddr},
		"1",
		false,
	)
	f := createPeersterProcess(
		"8005",
		fAddr,
		"F",
		[]string{},
		"1",
		false,
	)

	var aStdout bytes.Buffer
	a.Stdout = &aStdout
	err := a.Start()
	assert.NoError(t, err)
	err = b.Start()
	assert.NoError(t, err)
	err = c.Start()
	assert.NoError(t, err)
	err = d.Start()
	assert.NoError(t, err)
	err = e.Start()
	assert.NoError(t, err)
	err = f.Start()
	assert.NoError(t, err)

	// C (hits file1)
	// ^
	// B (hits file1) <- A   F (hits file2)
	// v                     ^
	// D----------E----------+

	// B's file
	_, _ = generateAndIndexFileWithSize(t, ChunkSize+100, "8001", "irrelevant_file_1")
	_, _ = generateAndIndexFileWithSize(t, ChunkSize+100, "8001", "irrelevant_file_2")
	file1MetaHash, file1Data := generateAndIndexFileWithSize(t, 3*ChunkSize, "8001", "ms word")

	// C's file
	_, _ = generateAndIndexFileWithSize(t, ChunkSize+100, "8002", "irrelevant_file_3")
	indexFile(t, file1Data, "8002", "aaa.txt")

	// F's file
	file2MetaHash, _ := generateAndIndexFileWithSize(t, 4*ChunkSize+200, "8005", "catapult inferior siege weapon")

	// Perform the search
	time.Sleep(2 * time.Second)
	sendCommand(t, "8000", "", "", "", "", []string{"cat", "ms", "aaa"}, "")
	time.Sleep(6 * time.Second)

	// Now we must download the files
	sendCommand(t,
		"8000", "", "something something.avi",
		"", fmt.Sprintf("%x", file1MetaHash), []string{}, "",
	)
	sendCommand(t,
		"8000", "", "indeed catapults are inferior siege weapon.txt",
		"", fmt.Sprintf("%x", file2MetaHash), []string{}, "",
	)
	time.Sleep(4 * time.Second)

	// Time to check the results
	err = a.Process.Kill()
	assert.NoError(t, err)
	err = b.Process.Kill()
	assert.NoError(t, err)
	err = c.Process.Kill()
	assert.NoError(t, err)
	err = d.Process.Kill()
	assert.NoError(t, err)
	err = e.Process.Kill()
	assert.NoError(t, err)
	err = f.Process.Kill()
	assert.NoError(t, err)

	aOutput := aStdout.String()
	fmt.Printf(aOutput)

	// Checking file search output
	assert.Contains(t,
		aOutput,
		fmt.Sprintf("FOUND match ms word at B metafile=%x chunks=1,2,3\n", file1MetaHash),
	)
	assert.Contains(t,
		aOutput,
		fmt.Sprintf("FOUND match aaa.txt at C metafile=%x chunks=1,2,3\n", file1MetaHash),
	)
	assert.Contains(t,
		aOutput,
		fmt.Sprintf("FOUND match catapult inferior siege weapon at F metafile=%x chunks=1,2,3,4,5\n", file2MetaHash),
	)
	assert.Contains(t, aOutput, "SEARCH FINISHED\n")

	// Checking file download output
	assert.Condition(t,
		func() (success bool) {
			return strings.Contains(aOutput, "DOWNLOADING metafile of something something.avi from B") ||
				strings.Contains(aOutput, "DOWNLOADING metafile of something something.avi from C")
		},
	)
	assert.Contains(t,
		aOutput,
		"DOWNLOADING metafile of indeed catapults are inferior siege weapon.txt from F",
	)
	for i := 0; i < 3; i++ {
		assert.Condition(t, func() (success bool) {
			return strings.Contains(aOutput, fmt.Sprintf("DOWNLOADING something something.avi chunk %d from B", i+1)) ||
				strings.Contains(aOutput, fmt.Sprintf("DOWNLOADING something something.avi chunk %d from C", i+1))
		})
	}
	for i := 0; i < 5; i++ {
		assert.Contains(t,
			aOutput,
			fmt.Sprintf("DOWNLOADING indeed catapults are inferior siege weapon.txt chunk %d from F", i+1),
		)
	}
	assert.Contains(t,
		aOutput,
		"RECONSTRUCTED file something something.avi",
	)
	assert.Contains(t,
		aOutput,
		"RECONSTRUCTED file indeed catapults are inferior siege weapon.txt",
	)

	// Checking that the downloaded files correspond to the original ones
	diff := exec.Command("diff", "_SharedFiles/catapult inferior siege weapon", "_Downloads/indeed catapults are inferior siege weapon.txt")
	err = diff.Run()
	assert.NoError(t, err) // If err is not nil, the files are different

	diff = exec.Command("diff", "_SharedFiles/ms word", "_Downloads/something something.avi")
	err = diff.Run()
	assert.NoError(t, err)

	diff = exec.Command("diff", "_SharedFiles/aaa.txt", "_Downloads/something something.avi")
	err = diff.Run()
	assert.NoError(t, err) // Should be exactly the same result as the one above, but we never know...
}

func TestPeerster_BlockchainBattle(t *testing.T) {
	var allFiles []string

	defer func() {
		for _, file := range allFiles {
			_ = os.Remove(file)
		}
	}()

	compile(t)

	// D <- C
	//      ^
	// A -> B -> E
	// v---------^

	aAddr := "localhost:3000"
	bAddr := "localhost:3001"
	cAddr := "localhost:3002"
	dAddr := "localhost:3003"
	eAddr := "localhost:3004"

	a := createPeersterProcess(
		"9000",
		aAddr,
		"A",
		[]string{bAddr, eAddr},
		"1",
		false,
	)
	b := createPeersterProcess(
		"9001",
		bAddr,
		"B",
		[]string{cAddr, eAddr},
		"1",
		false,
	)
	c := createPeersterProcess(
		"9002",
		cAddr,
		"C",
		[]string{dAddr},
		"1",
		false,
	)
	d := createPeersterProcess(
		"9003",
		dAddr,
		"D",
		[]string{},
		"1",
		false,
	)
	e := createPeersterProcess(
		"9004",
		eAddr,
		"E",
		[]string{},
		"1",
		false,
	)

	var aStdout bytes.Buffer
	var bStdout bytes.Buffer
	var cStdout bytes.Buffer
	var dStdout bytes.Buffer
	var eStdout bytes.Buffer
	a.Stdout = &aStdout
	b.Stdout = &bStdout
	c.Stdout = &cStdout
	d.Stdout = &dStdout
	e.Stdout = &eStdout

	err := a.Start()
	assert.NoError(t, err)
	err = b.Start()
	assert.NoError(t, err)
	err = c.Start()
	assert.NoError(t, err)
	err = d.Start()
	assert.NoError(t, err)
	err = e.Start()
	assert.NoError(t, err)

	nbFiles := uint32(128)

	for i := uint32(0); i < nbFiles; i++ {
		fileA := fmt.Sprintf("file_A_%d", i)
		fileB := fmt.Sprintf("file_B_%d", i)
		fileC := fmt.Sprintf("file_C_%d", i)
		fileD := fmt.Sprintf("file_D_%d", i)
		fileE := fmt.Sprintf("file_E_%d", i)

		allFiles = append(
			allFiles,
			"_SharedFiles/"+fileA,
			"_SharedFiles/.meta/"+fileA,

			"_SharedFiles/"+fileB,
			"_SharedFiles/.meta/"+fileB,

			"_SharedFiles/"+fileC,
			"_SharedFiles/.meta/"+fileC,

			"_SharedFiles/"+fileD,
			"_SharedFiles/.meta/"+fileD,

			"_SharedFiles/"+fileE,
			"_SharedFiles/.meta/"+fileE,
		)

		go generateAndIndexFile(t, "9000", fileA)
		go generateAndIndexFile(t, "9001", fileB)
		go generateAndIndexFile(t, "9002", fileC)
		go generateAndIndexFile(t, "9003", fileD)
		go generateAndIndexFile(t, "9004", fileE)
	}

	time.Sleep(10 * time.Second)

	// Time to check the results
	err = a.Process.Kill()
	assert.NoError(t, err)
	err = b.Process.Kill()
	assert.NoError(t, err)
	err = c.Process.Kill()
	assert.NoError(t, err)
	err = d.Process.Kill()
	assert.NoError(t, err)
	err = e.Process.Kill()
	assert.NoError(t, err)

	aOutput := aStdout.String()
	bOutput := bStdout.String()
	cOutput := cStdout.String()
	dOutput := dStdout.String()
	eOutput := eStdout.String()

	aChainUpd := findLastChainUpdate(aOutput)
	bChainUpd := findLastChainUpdate(bOutput)
	cChainUpd := findLastChainUpdate(cOutput)
	dChainUpd := findLastChainUpdate(dOutput)
	eChainUpd := findLastChainUpdate(eOutput)
	assert.NotEmpty(t, aChainUpd)
	assert.NotEmpty(t, bChainUpd)
	assert.NotEmpty(t, cChainUpd)
	assert.NotEmpty(t, dChainUpd)
	assert.NotEmpty(t, eChainUpd)

	largestChain := aChainUpd
	if len(bChainUpd) > len(largestChain) {
		largestChain = bChainUpd
	}
	if len(cChainUpd) > len(largestChain) {
		largestChain = cChainUpd
	}
	if len(dChainUpd) > len(largestChain) {
		largestChain = dChainUpd
	}
	if len(eChainUpd) > len(largestChain) {
		largestChain = eChainUpd
	}

	assert.Contains(t, largestChain, aChainUpd[len("CHAIN "):])
	assert.Contains(t, largestChain, bChainUpd[len("CHAIN "):])
	assert.Contains(t, largestChain, cChainUpd[len("CHAIN "):])
	assert.Contains(t, largestChain, dChainUpd[len("CHAIN "):])
	assert.Contains(t, largestChain, eChainUpd[len("CHAIN "):])
}

/*==== Helpers ====*/

func findLastChainUpdate(output string) string {
	lines := strings.Split(output, "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		line := lines[i]
		if strings.HasPrefix(line, "CHAIN") {
			return line
		}
	}
	return ""
}

func generateAndIndexFile(t *testing.T, gperUiPort, filename string) {
	fileSize := uint32(rand.Intn(int(2*ChunkSize)) + 100)
	_, _, data := GenerateRandomBytes(fileSize, ChunkSize)
	indexFile(t, data, gperUiPort, filename)
}

func generateAndIndexFileWithSize(t *testing.T, fileSize uint32, gperUiPort, filename string) (metaHash [sha256.Size]byte, data []byte) {
	_, metaHash, data = GenerateRandomBytes(fileSize, ChunkSize)
	indexFile(t, data, gperUiPort, filename)
	return
}

func indexFile(t *testing.T, data []byte, gperUiPort, filename string) {
	err := WriteTo("_SharedFiles", filename, data)
	assert.NoError(t, err)
	sendCommand(t, gperUiPort, "", filename, "", "", []string{}, "")
}

func sendCommand(
	t *testing.T,
	uiPort string,
	dest string,
	file string,
	message string,
	request string,
	keywords []string,
	budget string,
) {
	var args []string
	if uiPort != "" {
		args = append(args, "-UIPort="+uiPort)
	}
	if dest != "" {
		args = append(args, "-dest="+dest)
	}
	if file != "" {
		args = append(args, "-file="+file)
	}
	if message != "" {
		args = append(args, "-msg="+message)
	}
	if request != "" {
		args = append(args, "-request="+request)
	}
	if len(keywords) != 0 {
		args = append(args, "-keywords="+strings.Join(keywords, ","))
	}
	if budget != "" {
		args = append(args, "-budget="+budget)
	}

	cmd := exec.Command("./client", args...)
	cmd.Dir = "client"
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	assert.NoError(t, err)
	clientOutput := out.String()
	if strings.TrimSpace(clientOutput) != "" {
		fmt.Printf("Client output: %s\n", clientOutput)
	}
}

func createPeersterProcess(
	uiPort string,
	gossipAddr string,
	name string,
	peers []string,
	rtimer string,
	simple bool,
) *exec.Cmd {
	var args []string
	if uiPort != "" {
		args = append(args, "-UIPort="+uiPort)
	}
	if gossipAddr != "" {
		args = append(args, "-gossipAddr="+gossipAddr)
	}
	if name != "" {
		args = append(args, "-name="+name)
	}
	if len(peers) != 0 {
		args = append(args, "-peers="+strings.Join(peers, ","))
	}
	if rtimer != "" {
		args = append(args, "-rtimer="+rtimer)
	}
	if simple {
		args = append(args, "-simple")
	}
	cmd := exec.Command("./Peerster", args...)
	return cmd
}

func compile(t *testing.T) {
	buildOutput, err := doCompile("")
	assert.NoError(t, err)
	assert.Empty(t, buildOutput)
	buildOutput, err = doCompile("client")
	assert.NoError(t, err)
	assert.Empty(t, buildOutput)
}

func doCompile(dir string) (output string, err error) {
	cmd := exec.Command("go", "build")
	cmd.Dir = dir
	var out bytes.Buffer
	cmd.Stdout = &out
	err = cmd.Run()
	output = out.String()
	return
}

func WriteTo(directory, filename string, data []byte) error {
	file, err := os.Create(directory + "/" + filename)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(data)
	return err
}

func GenerateRandomBytes(length uint32, chunkSize uint32) (metaFileBytes []byte, metaHash [sha256.Size]byte, data []byte) {
	nbChunks := IntCeilDiv(length, chunkSize)
	data = make([]byte, length)
	rand.Read(data)

	metaFileBytes = make([]byte, nbChunks*sha256.Size)
	for i := uint32(0); i < nbChunks; i++ {
		dataFrom := i * chunkSize
		dataUntil := (i + 1) * chunkSize
		if dataUntil > length {
			dataUntil = length
		}
		hash := sha256.Sum256(data[dataFrom:dataUntil])
		copy(metaFileBytes[i*sha256.Size:(i+1)*sha256.Size], hash[:])
	}
	metaHash = sha256.Sum256(metaFileBytes)
	return
}

// https://stackoverflow.com/a/2745086
func IntCeilDiv(x, y uint32) uint32 {
	if x == 0 {
		return 0
	} else {
		return 1 + ((x - 1) / y)
	}
}
