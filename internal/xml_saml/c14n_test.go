package xml_saml

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"testing"
)

type TestData struct {
	name      string
	canonType CanonicalizationType
}

func Test_Canonicalization(t *testing.T) {
	testDatas := []TestData{
		TestData{name: "c14n_001", canonType: XML_C14N},
		TestData{name: "c14n_002", canonType: XML_C14N},
	}

	for _, testData := range testDatas {
		inputFile, err := os.Open(fmt.Sprintf("data/%s_input.xml", testData.name))
		if err != nil {
			t.Errorf("Unable to open test input file: %T %v", err, err)
			continue
		}
		expectedBytes, err := ioutil.ReadFile(fmt.Sprintf("data/%s_expected.xml.", testData.name))
		if err != nil {
			t.Errorf("Unable to read expected data file: %T %v", err, err)
			continue
		}
		expectedData := string(expectedBytes[:len(expectedBytes)-1])

		root, err := Parse(bufio.NewReader(inputFile))
		if err != nil {
			t.Errorf("Unexpected error: %T %v", err, err)
			continue
		}

		c14nRoot, err := root.Canonicalize(testData.canonType)
		if err != nil {
			t.Errorf("Unexpected error: %T %v", err, err)
			continue
		}
		var b bytes.Buffer
		c14nRoot.Print(&b)
		output := b.String()
		if expectedData != output {
			t.Errorf("Canonicalization `%v` mismatched `%v`", output, expectedData)
		}
	}
}
