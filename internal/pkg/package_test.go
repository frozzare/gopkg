package pkg

import (
	util "github.com/go-pkg-org/gopkg/internal/util"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestCreateEntries(t *testing.T) {
	dir, _ := ioutil.TempDir("", "gopkg_*")
	t.Cleanup(func() {
		os.RemoveAll(dir)
	})

	jsonFile, _ := ioutil.TempFile(dir, "*.json")
	txtFile, _ := ioutil.TempFile(dir, "*.txt")
	xmlFile, _ := ioutil.TempFile(dir, "*.xml")

	result, _ := CreateEntries(dir, "some/prefix", []string{})

	expectedPaths := []string{
		filepath.Join(jsonFile.Name()),
		filepath.Join(txtFile.Name()),
		filepath.Join(xmlFile.Name()),
	}
	expectedArchivePaths := []string{
		filepath.Join("some", "prefix", filepath.Base(txtFile.Name())),
		filepath.Join("some", "prefix", filepath.Base(jsonFile.Name())),
		filepath.Join("some", "prefix", filepath.Base(xmlFile.Name())),
	}

	if len(result) != len(expectedPaths) {
		t.Error("length mismatch between expected and result")
	}

	for _, f := range result {
		if !util.Contains(expectedPaths, f.FilePath) {
			t.Errorf("%s did not exist in expected paths", f.FilePath)
		}
		if !util.Contains(expectedArchivePaths, f.ArchivePath) {
			t.Errorf("%s did not exist in expected archive paths", f.ArchivePath)
		}
	}

	result, _ = CreateEntries(dir, "some/prefix", []string{filepath.Base(xmlFile.Name())})
	expectedPaths = []string{
		filepath.Join(jsonFile.Name()),
		filepath.Join(txtFile.Name()),
	}
	expectedArchivePaths = []string{
		filepath.Join("some", "prefix", filepath.Base(txtFile.Name())),
		filepath.Join("some", "prefix", filepath.Base(jsonFile.Name())),
	}

	if len(result) != len(expectedPaths) {
		t.Errorf("length mismatch between expected and result (got %d want %d)", len(result), len(expectedPaths))
	}

	for _, f := range result {
		if !util.Contains(expectedPaths, f.FilePath) {
			t.Errorf("%s did not exist in expected paths", f.FilePath)
		}
		if !util.Contains(expectedArchivePaths, f.ArchivePath) {
			t.Errorf("%s did not exist in expected archive paths", f.ArchivePath)
		}
	}
}

func TestWrite(t *testing.T) {
	dir, _ := ioutil.TempDir("", "gopkg_*")
	t.Cleanup(func() {
		os.RemoveAll(dir)
	})

	jsonFile, _ := ioutil.TempFile(dir, "*.json")
	txtFile, _ := ioutil.TempFile(dir, "*.txt")
	xmlFile, _ := ioutil.TempFile(dir, "*.xml")

	err := Write(filepath.Join(dir, "out.pkg"), []Entry{
		{xmlFile.Name(), "test/xmlfile.xml"},
		{jsonFile.Name(), "jsonfile.json"},
		{txtFile.Name(), "txtfile.txt"},
	}, true)

	if err != nil {
		t.Errorf("failed to create archive: %s", err)
	}

	if _, err := os.Stat(filepath.Join(dir, "out.pkg")); err != nil {
		t.Errorf("archive created was not written to disk: %s", err)
	}
}

func TestRead(t *testing.T) {
	dir, _ := ioutil.TempDir("", "gopkg_*")
	t.Cleanup(func() {
		os.RemoveAll(dir)
	})

	jsonFile, _ := ioutil.TempFile(dir, "*.json")
	jsonFile.WriteString("This is a json file")
	jsonFile.Close()
	txtFile, _ := ioutil.TempFile(dir, "*.txt")
	txtFile.WriteString("This is a txt file")
	txtFile.Close()
	xmlFile, _ := ioutil.TempFile(dir, "*.xml")
	xmlFile.WriteString("This is an xml file")
	xmlFile.Close()

	err := Write(filepath.Join(dir, "out.pkg"), []Entry{
		{xmlFile.Name(), "test/xmlfile.xml"},
		{jsonFile.Name(), "jsonfile.json"},
		{txtFile.Name(), "txtfile.txt"},
	}, true)

	if err != nil {
		t.Errorf("failed to create archive: %s", err)
	}

	list, err := Read(filepath.Join(dir, "out.pkg"))
	if err != nil {
		t.Errorf("failed to read the archive: %s", err)
	}

	jsonContent := string(list["jsonfile.json"])
	xmlContent := string(list["test/xmlfile.xml"])
	txtContent := string(list["txtfile.txt"])

	if jsonContent != "This is a json file" {
		t.Errorf("Json file could not be read.")
	}
	if xmlContent != "This is an xml file" {
		t.Errorf("Xml file could not be read.")
	}
	if txtContent != "This is a txt file" {
		t.Errorf("Txt file could not be read.")
	}

}

func TestGetName(t *testing.T) {
	name, err := GetName("github-creekorful-mvnparser", "1.0.0", "", "", Control)
	if err != nil {
		t.Error(err)
	}

	if name != "github-creekorful-mvnparser_1.0.0.pkg" {
		t.Errorf("wrong package name (%s)", name)
	}

	name, err = GetName("github-creekorful-mvnparser", "1.0.0", "", "", Source)
	if err != nil {
		t.Error(err)
	}

	if name != "github-creekorful-mvnparser_1.0.0-dev.pkg" {
		t.Errorf("wrong package name (%s)", name)
	}

	name, err = GetName("gohello", "1.0.0", "linux", "amd64", Binary)
	if err != nil {
		t.Error(err)
	}

	if name != "gohello_1.0.0_linux_amd64.pkg" {
		t.Error(err)
	}
}

func TestParseName(t *testing.T) {
	name, ver, _, _, pkgType, err := ParseName("github-creekorful-mvnparser_1.0.0.pkg")
	if err != nil {
		t.Error(err)
	}

	if name != "github-creekorful-mvnparser" {
		t.Error("wrong package name")
	}

	if ver != "1.0.0" {
		t.Error("wrong package version")
	}

	if pkgType != Control {
		t.Error("package should be control")
	}

	name, ver, _, _, pkgType, err = ParseName("github-creekorful-mvnparser_1.0.0-dev.pkg")
	if err != nil {
		t.Error(err)
	}

	if name != "github-creekorful-mvnparser" {
		t.Error("wrong package name")
	}

	if ver != "1.0.0" {
		t.Error("wrong package version")
	}

	if pkgType != Source {
		t.Error("package should be source")
	}

	name, ver, o, arch, pkgType, err := ParseName("gohello_1.0.0_linux_amd64.pkg")
	if err != nil {
		t.Error(err)
	}

	if name != "gohello" {
		t.Error("wrong package name")
	}

	if ver != "1.0.0" {
		t.Error("wrong package version")
	}

	if o != "linux" {
		t.Error("wrong os")
	}

	if arch != "amd64" {
		t.Error("wrong arch")
	}

	if pkgType != Binary {
		t.Error("package should be binary")
	}
}
