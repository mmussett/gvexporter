package main

import (
	"archive/zip"
	_ "encoding/base64"
	"encoding/xml"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
)

type Data struct {
}

type DeploymentDescriptorFactory struct {
	Text                                 string `xml:",chardata"`
	Name                                 string `xml:"name"`
	RequiresConfiguration                string `xml:"requiresConfiguration"`
	DeploymentDescriptorFactoryClassName string `xml:"deploymentDescriptorFactoryClassName"`
}

type NameValuePairs struct {
	Text                  string `xml:",chardata"`
	Name                  string `xml:"name"`
	RequiresConfiguration string `xml:"requiresConfiguration"`
	NameValuePair         []struct {
		Text                  string `xml:",chardata"`
		Name                  string `xml:"name"`
		RequiresConfiguration string `xml:"requiresConfiguration"`
		Value                 string `xml:"value"`
	} `xml:"NameValuePair"`
}

type RepoInstance struct {
	Text                         string `xml:",chardata"`
	Repoinstance                 string `xml:"repoinstance,attr"`
	Name                         string `xml:"name"`
	RequiresConfiguration        string `xml:"requiresConfiguration"`
	DisableConfigureAtDeployment string `xml:"disableConfigureAtDeployment"`
}

type Modules struct {
	Text                         string `xml:",chardata"`
	Name                         string `xml:"name"`
	RequiresConfiguration        string `xml:"requiresConfiguration"`
	DisableConfigureAtDeployment string `xml:"disableConfigureAtDeployment"`
	PathName                     string `xml:"pathName"`
}

type DeploymentDescriptors struct {
	XMLName                     xml.Name                      `xml:"DeploymentDescriptors"`
	Text                        string                        `xml:",chardata"`
	Xmlns                       string                        `xml:"xmlns,attr"`
	Name                        string                        `xml:"name"`
	Description                 string                        `xml:"description"`
	Version                     string                        `xml:"version"`
	Owner                       string                        `xml:"owner"`
	CreationDate                string                        `xml:"creationDate"`
	IsApplicationArchive        string                        `xml:"isApplicationArchive"`
	DeploymentDescriptorFactory []DeploymentDescriptorFactory `xml:"DeploymentDescriptorFactory"`
	RepoInstance                []RepoInstance                `xml:"RepoInstance"`
	NameValuePairs              []NameValuePairs              `xml:"NameValuePairs"`
	Modules                     []Modules                     `xml:"Modules"`
}

func main() {

	earFilename := flag.String("ear", "", "Input EAR Filename")
	outputFilename := flag.String("o", "", "Output Filename")
	flag.Parse()

	if *earFilename == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	if *outputFilename == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	log.Printf("Extracting GV from EAR File %s to %s\n", *earFilename, *outputFilename)

	zr, err := zip.OpenReader(*earFilename)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	// Extract TIBCO.xml from EAR
	found := false
	for _, file := range zr.File {
		if file.Name != "TIBCO.xml" {
			continue
		} else {
			log.Println("Found TIBCO.xml in EAR, extracting.")
			zipFile, err := file.Open()
			if err != nil {
				log.Fatal(err)
			}

			tmpFile, err := os.OpenFile("./TIBCO.xml", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
			if err != nil {
				log.Fatal(err)
			}

			_, err = io.Copy(tmpFile, zipFile)
			if err != nil {
				log.Fatal(err)
			}

			zipFile.Close()
			found = true
			break
		}
	}

	if !found {
		log.Fatal("Unable to find TIBCO.xml in EAR file")
	}

	inFile, err := os.OpenFile("./TIBCO.xml", os.O_RDONLY, 0)
	if err != nil {
		log.Fatal(err)
	}

	os.Remove(*outputFilename)
	os.Remove("./TIBCO.xml")

	outFile, err := os.OpenFile(*outputFilename, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println(err)
		panic(-1)
	}

	defer inFile.Close()
	defer outFile.Close()

	// Unmarshal our XML

	log.Println("Exporting GVs as JSON")

	byteValue, _ := ioutil.ReadAll(inFile)

	dd := DeploymentDescriptors{}
	xml.Unmarshal(byteValue, &dd)

	outFile.WriteString("{\n")
	for _, nvp := range dd.NameValuePairs[0].NameValuePair {
		key := nvp.Name
		value := nvp.Value
		outString := fmt.Sprintf("  %q : %q\n", key, value)
		outFile.WriteString(outString)
	}
	outFile.WriteString("}\n")

	log.Println("Done")

}
