package ini

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

var (
	complexString = `   ;comment1 
	defkey1= defvalue1
	defkey2=
defkey3 = defvalue3 #comment2

  [   section1   ] #seccomment1
      key1 = value1  
key2 = value2# comment3

          # comment4

[section2 ]

[SECTION3] #seccomment2
s3key1 =
s3key2 = s3value2      # comment5

[ 毚饯襃ブみょ ]
䥵妦飌ぞ盯 = 覎びゅフォ駧橜 槞㨣

[test2]
test=foo
[test3]

test=bar
	`
)

func TestParseBad(t *testing.T) {
	_, err := LoadString("bad")
	if err == nil {
		t.Error("bad string should error")
	}

	r := strings.NewReader("bad")
	_, err = Load(r)
	if err == nil {
		t.Error("bad string should error")
	}
}

func TestParseEmpty(t *testing.T) {
	f, err := LoadString(``)
	if err != nil {
		t.Error("could not load blank string")
	}

	d0 := "\n"
	if d0 != f.String() {
		t.Error("new line should be added to blank string")
	}
}

func TestParseComplex(t *testing.T) {
	f, err := LoadString(complexString)
	if err != nil {
		t.Error("could not load complexString")
	}

	// test raw section name parsing
	eraw := []string{"", "   section1   ", "section2 ", "SECTION3", " 毚饯襃ブみょ "}
	rawnames := f.RawSectionNames()
	for i, r := range eraw {
		if r != rawnames[i] {
			t.Errorf("raw section name %d should be '%s', got: '%s'", i, r, rawnames[i])
		}
	}

	// test clean section name parsing
	enam := []string{"", "section1", "section2", "section3", "毚饯襃ブみょ"}
	names := f.SectionNames()
	for i, n := range enam {
		if n != names[i] {
			t.Errorf("section name %d should be '%s', got: '%s'", i, n, names[i])
		}
	}
}

func TestGetSectionAndSectionKeys(t *testing.T) {
	f, err := LoadString(complexString)
	if err != nil {
		t.Error("could not load complexString")
	}

	// check for nonexistent section name
	sect := f.GetSection("nonexistent")
	if sect != nil {
		t.Error("GetSection should return nil for nonexistent section")
	}

	sectionkeys := map[string][]string{
		"":         {"defkey1", "defkey2", "defkey3"},
		"section1": {"key1", "key2"},
		"section2": {},
		"section3": {"s3key1", "s3key2"},
		"毚饯襃ブみょ":   {"䥵妦飌ぞ盯"},
		"test2":    {"test"},
		"test3":    {"test"},
	}

	// check sections and key combinations
	for name, keys := range sectionkeys {
		// check section is present
		s := f.GetSection(name)
		if s == nil {
			t.Errorf("Section '%s' should be present", name)
		}

		// make sure section name is same
		if name != s.Name() {
			t.Errorf("Section '%s' should have same name as Section.Name()", name)
		}

		// compare section keys
		kys := s.Keys()
		for i, k := range keys {
			if k != kys[i] {
				t.Errorf("Section '%s' should have key '%s' (%d), got: '%s'", name, k, i, kys[i])
			}
		}
	}
}

func TestRawSectionKeys(t *testing.T) {
	f, err := LoadString(complexString)
	if err != nil {
		t.Error("could not load complexString")
	}

	rawkeys := map[string][]string{
		"":         {"defkey1", "defkey2", "defkey3 "},
		"section1": {"key1 ", "key2 "},
		"section2": {},
		"section3": {"s3key1 ", "s3key2 "},
		"毚饯襃ブみょ":   {"䥵妦飌ぞ盯 "},
		"test2":    {"test"},
		"test3":    {"test"},
	}

	// check sections and key combinations
	for name, keys := range rawkeys {
		// check section is present
		s := f.GetSection(name)
		if s == nil {
			t.Errorf("Section '%s' should be present", name)
		}

		// make sure section name is same
		if name != s.Name() {
			t.Errorf("Section '%s' should have same name as Section.Name()", name)
		}

		// compare section keys
		kys := s.RawKeys()
		for i, k := range keys {
			if k != kys[i] {
				t.Errorf("Section '%s' should have key '%s' (%d), got: '%s'", name, k, i, kys[i])
			}
		}
	}
}

func TestGetKey(t *testing.T) {
	f, err := LoadString(complexString)
	if err != nil {
		t.Error("could not load complexString")
	}

	keyvaluemap := map[string]string{
		"defkey1":         "defvalue1",
		"defkey2":         "",
		"defkey3":         "defvalue3",
		"defvalue4":       "",
		"section1":        "",
		"section1.key1":   "value1",
		"SECTION1.KEY1":   "value1",
		"section2.key":    "",
		"section3.s3key1": "",
		"section3.s3key2": "s3value2",

		"毚饯襃ブみょ.䥵妦飌ぞ盯": "覎びゅフォ駧橜 槞㨣",

		"test2.test":              "foo",
		"test3.test":              "bar",
		"nonexistent.nonexistent": "",
	}

	// check each key value as per map
	for key, exp := range keyvaluemap {
		val := f.GetKey(key)
		if exp != val {
			t.Errorf("for key '%s' expected '%s', got: '%s'", key, exp, val)
		}
	}
}

// test preservation of whitespace
func TestPreservation(t *testing.T) {
	d0 := " #com1  \n	[sect1 ] ;com2\n  k1 = v1  ;com3\n  k2=   \n  k3 = v3\n\n  [ sect2 ]\n  [sect3]\n  k4= v4 \n	  \n"
	f, err := LoadString(d0)
	if err != nil {
		t.Error("could not load string")
	}

	// set some keys
	f.SetKey("sect1.k1", "v0")
	f.SetKey("sect1.k2", "v5")
	f.SetKey("sect3.k4", "")

	// check whitespace has been preserved
	d1 := " #com1  \n	[sect1 ] ;com2\n  k1 = v0;com3\n  k2=   v5\n  k3 = v3\n\n  [ sect2 ]\n  [sect3]\n  k4= \n	  \n"
	if d1 != f.String() {
		t.Error("SetKey should preserve all spacing and comments")
	}
}

func TestAddSection(t *testing.T) {
	f, err := LoadString("")
	if err != nil {
		t.Error("could not load blank string")
	}

	f.AddSection("awesome")
	d1 := "[awesome]\n"
	if d1 != f.String() {
		t.Errorf("strings should match: '%s'", f)
	}

	f.AddSection("second")
	d2 := "[awesome]\n[second]\n"
	if d2 != f.String() {
		t.Errorf("strings should match: '%s'", f)
	}

	f.AddSection("") // should have no effect
	if d2 != f.String() {
		t.Errorf("strings should match: '%s'", f)
	}
}

func TestRenameSection(t *testing.T) {
	d1 := " #com1  \n	[sect1 ] ;com2\n  k1 = v1  ;com3\n  k2=   \n  k3 = v3\n\n  [ sect2 ]\n  [sect3]\n  k4= v4 \n	  \n"
	f, err := LoadString(d1)
	if err != nil {
		t.Error("could not load string")
	}

	// verify that setkey works first
	f.SetKey("sect1.k1", "v0")
	d2 := " #com1  \n	[sect1 ] ;com2\n  k1 = v0;com3\n  k2=   \n  k3 = v3\n\n  [ sect2 ]\n  [sect3]\n  k4= v4 \n	  \n"
	if d2 != f.String() {
		t.Error("SetKey should preserve location, spacing, and comments")
	}

	// check basic rename section
	f.RenameSection("sect1", "sect4")
	d3 := " #com1  \n	[sect4] ;com2\n  k1 = v0;com3\n  k2=   \n  k3 = v3\n\n  [ sect2 ]\n  [sect3]\n  k4= v4 \n	  \n"
	if d3 != f.String() {
		t.Error("RenameSection should preserve location, spacing, and comments")
	}

	// check section names
	enames := []string{"", "sect4", "sect2", "sect3"}
	names := f.SectionNames()
	for i, n := range enames {
		if n != names[i] {
			t.Error("after RenameSection, SectionNames should be in preserved order")
		}
	}

	// check getkey
	val := f.GetKey("sect4.k1")
	if val != "v0" {
		t.Errorf("after RenameSection, GetKey should correctly return key values from renamed section, got: '%s'", val)
	}

	// test that sect0 is gone
	s := f.GetSection("sect0")
	if s != nil {
		t.Error("after RenameSection, sect0 should no longer be defined")
	}

	f.SetKey("sect4.k2", "foobar")
	d4 := " #com1  \n	[sect4] ;com2\n  k1 = v0;com3\n  k2=   foobar\n  k3 = v3\n\n  [ sect2 ]\n  [sect3]\n  k4= v4 \n	  \n"
	if d4 != f.String() {
		t.Error("after RenameSection, SetKey should preserve location, spacing, and comments")
	}
}

func TestSetKey(t *testing.T) {
	f, err := LoadString("")
	if err != nil {
		t.Error("could not load blank string")
	}

	f.SetKey("sect1.key1", "val1")
	d1 := "[sect1]\n\tkey1=val1\n"
	if d1 != f.String() {
		t.Error("SetKey should set a value correctly")
	}

	f.SetKey("sect2.key2", "val2")
	d2 := "[sect1]\n\tkey1=val1\n[sect2]\n\tkey2=val2\n"
	if d2 != f.String() {
		t.Error("SetKey should set a value correctly")
	}

	f.SetKey("key0", "val0")
	d3 := "key0=val0\n[sect1]\n\tkey1=val1\n[sect2]\n\tkey2=val2\n"
	if d3 != f.String() {
		t.Error("SetKey should set a value correctly")
	}
}

// check that key names are being retrived and set from the correct sections,
// even if having same key
func TestSameKey(t *testing.T) {
	d0 := "[sect1]\nkey=val1\n[sect2]\nkey=val2\n"
	f, err := LoadString(d0)
	if err != nil {
		t.Error("could not load string")
	}

	k1 := f.GetKey("sect1.key")
	if k1 != "val1" {
		t.Error("sect1.key should be val1")
	}

	k2 := f.GetKey("sect2.key")
	if k2 != "val2" {
		t.Error("sect2.key should be val2")
	}

	f.SetKey("sect1.key", "val3")
	d1 := "[sect1]\nkey=val3\n[sect2]\nkey=val2\n"
	if d1 != f.String() {
		t.Error("SetKey should set a value correctly")
	}

	f.SetKey("sect2.key", "val4")
	d2 := "[sect1]\nkey=val3\n[sect2]\nkey=val4\n"
	if d2 != f.String() {
		t.Error("SetKey should set a value correctly")
	}

	k3 := f.GetKey("sect1.key")
	if k3 != "val3" {
		t.Error("sect1.key should be val3")
	}

	k4 := f.GetKey("sect2.key")
	if k4 != "val4" {
		t.Error("sect4.key should be val4")
	}
}

func TestSetKeyAdvanced(t *testing.T) {
	d0 := "k0=val0\n\n\n[sect1]\nk1=v1\n\nk2=v2\nk3=v3\n\n\n[sect5]\nk6=val6\n"
	f, err := LoadString(d0)
	if err != nil {
		t.Error("could not load string")
	}

	f.SetKey("k1", "val1")
	d1 := "k0=val0\nk1=val1\n\n\n[sect1]\nk1=v1\n\nk2=v2\nk3=v3\n\n\n[sect5]\nk6=val6\n"
	if d1 != f.String() {
		t.Error("SetKey should correctly insert a key after last blank line of empty section")
	}

	f.SetKey("sect1.k4", "val4")
	d2 := "k0=val0\nk1=val1\n\n\n[sect1]\nk1=v1\n\nk2=v2\nk3=v3\nk4=val4\n\n\n[sect5]\nk6=val6\n"
	if d2 != f.String() {
		t.Error("SetKey should correctly insert a key after last blank line of section sect1 (copied whitespace)")
	}

	f.SetKey("k7", "val7")
	d3 := "k0=val0\nk1=val1\nk7=val7\n\n\n[sect1]\nk1=v1\n\nk2=v2\nk3=v3\nk4=val4\n\n\n[sect5]\nk6=val6\n"
	if d3 != f.String() {
		t.Error("SetKey should correctly insert a key after last blank line of default section")
	}
}

func TestRemoveSection(t *testing.T) {
	d0 := "[sect0]\nkey1=val1\n[sect1]\n[sect3]\nkey2=val2\n[sect4]\n"
	f, err := LoadString(d0)
	if err != nil {
		t.Error("could not load string")
	}

	// test removing nonexistent
	f.RemoveSection("nonexistent")
	if d0 != f.String() {
		t.Error("after RemoveSection for nonexistent section, data should be the same as original")
	}

	// test remove section in middle with no keys
	f.RemoveSection("sect1")
	d1 := "[sect0]\nkey1=val1\n[sect3]\nkey2=val2\n[sect4]\n"
	if d1 != f.String() {
		t.Error("could not RemoveSection sect1")
	}

	// remove end section with no keys
	f.RemoveSection("sect4")
	d2 := "[sect0]\nkey1=val1\n[sect3]\nkey2=val2\n"
	if d2 != f.String() {
		t.Error("could not RemoveSection sect4")
	}

	// remove end section with keys
	f.RemoveSection("sect3")
	d3 := "[sect0]\nkey1=val1\n"
	if d3 != f.String() {
		t.Error("could not RemoveSection sect3")
	}

	// remove first section
	f.RemoveSection("sect0")
	d4 := "\n"
	if d4 != f.String() {
		t.Error("could not RemoveSection sect0")
	}

	// after, line count should still be 1
	if f.LineCount() != 1 {
		t.Error("after removing all sections, len(lines) should equal 1")
	}
}

func TestRemoveKey(t *testing.T) {
	d0 := "k0=val0\n[sect1]\nk1=val1\n\nk2=val2\n[sect2]\n"
	f, err := LoadString(d0)
	if err != nil {
		t.Error("could not load string")
	}

	// test set key first
	f.SetKey("sect2.k3", "val3")
	d1 := "k0=val0\n[sect1]\nk1=val1\n\nk2=val2\n[sect2]\n\tk3=val3\n"
	if d1 != f.String() {
		t.Error("SetKey should correctly add a key")
	}

	f.RemoveKey("sect2.k3")
	if d0 != f.String() {
		t.Error("RemoveKey should correctly remove a key")
	}

	f.RemoveKey("k0")
	d2 := "[sect1]\nk1=val1\n\nk2=val2\n[sect2]\n"
	if d2 != f.String() {
		t.Error("RemoveKey should correctly remove a key")
	}

	f.RemoveKey("sect1.k1")
	d3 := "[sect1]\n\nk2=val2\n[sect2]\n"
	if d3 != f.String() {
		t.Error("RemoveKey should correctly remove a key")
	}

	f.RemoveKey("sect1.k2")
	d4 := "[sect1]\n\n[sect2]\n"
	if d4 != f.String() {
		t.Error("RemoveKey should correctly remove a key")
	}

	f.RemoveKey("sect1.k2")
	f.RemoveKey("nonexistent.nonexistent")
	if d4 != f.String() {
		t.Error("RemoveKey called on nonexistent keys should not alter the file")
	}
}

// Test git style names (ie, subsections)
func TestGitStyleNames(t *testing.T) {
	d0 := "  [  sect0   \"sub0\"  ] ;comment \n  k0 = v0  \n"
	f, err := LoadString(d0)
	if err != nil {
		t.Error("could not load string")
	}

	// force Gitconfig style names
	f.SectionManipFunc = GitSectionManipFunc
	f.SectionNameFunc = GitSectionNameFunc

	enam := []string{"", "sect0.sub0"}
	names := f.SectionNames()
	for i, n := range enam {
		if n != names[i] {
			t.Errorf("section name %d should be '%s', got: '%s'", i, n, names[i])
		}
	}

	s := f.GetSection("sect0.nonexistent")
	if s != nil {
		t.Error("section sect0.nonexistent should not be present")
	}

	sect0 := f.GetSection("sect0.sub0")
	if sect0 == nil {
		t.Fatal("section sect0.sub0 should be present")
	}

	k0 := f.GetKey("sect0.sub0.k0")
	if k0 != "v0" {
		t.Error("sect0.sub0.k0 should be v0")
	}

	f.SetKey("sect0.sub0.k0", "v1")
	d1 := "  [  sect0   \"sub0\"  ] ;comment \n  k0 = v1\n"
	if d1 != f.String() {
		t.Error("setting key using GitSectionManipFunc should preserve location, spacing, and comments")
	}

	f.RenameSection("sect0.sub0", "sect1.sub1")
	d2 := "  [sect1 \"sub1\"] ;comment \n  k0 = v1\n"
	if d2 != f.String() {
		t.Error("after RenameSection using GitSectionManipFunc, location, spacing and comments when using should be preserved")
	}

	f.SetKey("sect2.sub2.k2", "v2")
	d3 := "  [sect1 \"sub1\"] ;comment \n  k0 = v1\n[sect2 \"sub2\"]\n\tk2=v2\n"
	if d3 != f.String() {
		t.Error("setting key using GitSectionManipFunc should correctly add a section, key")
	}

	sect2 := f.GetSection("sect2.sub2")
	if sect2 == nil || sect2.Name() != "sect2.sub2" {
		t.Error("sect2.sub2 should be defined and Name should be sect2.sub2")
	}

	v2 := f.GetKey("sect2.sub2.k2")
	if v2 != "v2" {
		t.Error("sect2.sub2.k2 value should be v2")
	}

	f.SetKey("sect3.k3", "v3")
	d4 := "  [sect1 \"sub1\"] ;comment \n  k0 = v1\n[sect2 \"sub2\"]\n\tk2=v2\n[sect3]\n\tk3=v3\n"
	if d4 != f.String() {
		t.Error("setting key using GitSectionManipFunc should correctly add a section, key")
	}

	v3 := f.GetKey("sect3.k3")
	if v3 != "v3" {
		t.Error("sect3.k3 should be v3")
	}

	f.SetKey("k4", "v4")
	d5 := "k4=v4\n  [sect1 \"sub1\"] ;comment \n  k0 = v1\n[sect2 \"sub2\"]\n\tk2=v2\n[sect3]\n\tk3=v3\n"
	if d5 != f.String() {
		t.Error("setting key using GitSectionManipFunc should correctly add key to default section")
	}

	v4 := f.GetKey("k4")
	if v4 != "v4" {
		t.Error("k4 should be v4")
	}

	f.SetKey("sect3.sect3.k4", "value4")
	d6 := "k4=v4\n  [sect1 \"sub1\"] ;comment \n  k0 = v1\n[sect2 \"sub2\"]\n\tk2=v2\n[sect3]\n\tk3=v3\n[sect3 \"sect3\"]\n\tk4=value4\n"
	if d6 != f.String() {
		t.Error("setting key using GitSectionManipFunc should correctly add section and key")
	}

	enam = []string{"", "sect1.sub1", "sect2.sub2", "sect3", "sect3.sect3"}
	names = f.SectionNames()
	for i, n := range enam {
		if n != names[i] {
			t.Errorf("section name %d should be '%s', got: '%s'", i, n, names[i])
		}
	}

	f.RemoveSection("sect3.sect3")
	if d5 != f.String() {
		t.Error("removing section sect3.sect3 should result in previous version of file")
	}
}

func TestStringValues(t *testing.T) {
	d0 := "k0=\"v0;#notacomment\"\nk1=\"line0\nline2\"\n"
	f, err := LoadString(d0)
	if err != nil {
		t.Error("could not load string")
	}

	v0 := f.GetKey("k0")
	if v0 != "\"v0;#notacomment\"" {
		t.Errorf("v0 should be \"v0;#notacomment\"")
	}

	v1 := f.GetKey("k1")
	if v1 != "\"line0\nline2\"" {
		t.Error("k1 should span multiple lines")
	}
}

func TestBadWrite(t *testing.T) {
	f := NewFile()
	f.SetKey("k1", "v1")
	err := f.File.Write("")
	if err == nil {
		t.Error("file should not be writable")
	}

	err = f.Save()
	if err == nil {
		t.Error("file should not be writable")
	}
}

func TestCustomSectionCompFunc(t *testing.T) {
	f := NewFile()
	f.SectionCompFunc = func(a, b string) bool {
		return true
	}

	f.SetKey("k1", "v1")
	v1 := f.GetKey("sect0.k1")
	if v1 != "v1" {
		t.Error("using custom section comp func, section names should be ignored and v1 should be v1")
	}
}

func TestMaps(t *testing.T) {
	tmpfile, err := ioutil.TempFile("", "")
	if err != nil {
		t.Fatal("could not create temporary file")
	}

	tmpfile.Close()

	defer func() {
		err := os.Remove(tmpfile.Name())
		if err != nil {
			t.Fatalf("could not remove temporary file %s", tmpfile.Name())
		}
	}()

	values := map[string]map[string]string{
		"": {
			"k0": "v0",
		},
		"sect0": {
			"k1": "v1",
			"k2": "v2",
		},
		"sect0.sub1": {
			"k3": "v3",
		},
	}

	// create file data
	f0 := NewFile()
	f0.Filename = tmpfile.Name()
	f0.SectionManipFunc = GitSectionManipFunc
	f0.SectionNameFunc = GitSectionNameFunc
	f0.SetMap(values)

	// test map
	map0 := f0.GetMap()
	for name, vals := range values {
		section := f0.GetSection(name)
		if section == nil {
			t.Fatalf("section %s should be present", name)
		}

		for key, val := range vals {
			v := map0[name][key]
			if val != v {
				t.Errorf("section %s should have key %s with value %s, got: '%s'", name, key, val, v)
			}
		}
	}

	// write to disk
	err = f0.Save()
	if err != nil {
		t.Fatalf("could not save tmpfile: %s", err)
	}

	// load data back
	f1, err := LoadFile(tmpfile.Name())
	f1.SectionManipFunc = GitSectionManipFunc
	f1.SectionNameFunc = GitSectionNameFunc
	if err != nil {
		t.Fatalf("could not open tmpfile: %s", err)
	}

	expvalues := map[string]string{
		"k0":            "v0",
		"sect0.k1":      "v1",
		"sect0.k2":      "v2",
		"sect0.sub1.k3": "v3",
	}

	for key, val := range expvalues {
		v := f1.GetKey(key)
		if val != v {
			t.Errorf("expected %s to be %s, got: '%s'", key, val, v)
		}
	}

	// compare flat map
	map1 := f1.GetMapFlat()
	for key, val := range expvalues {
		v, ok := map1[key]
		if !ok || val != v {
			t.Errorf("expected %s to be %s, got: '%s'", key, val, v)
		}
	}

	// set map using flat values
	f2 := NewFile()
	f2.SectionManipFunc = GitSectionManipFunc
	f2.SectionNameFunc = GitSectionNameFunc
	f2.SetMapFlat(expvalues)
	for key, val := range expvalues {
		v := f2.GetKey(key)
		if val != v {
			t.Errorf("expected %s to be %s, got: '%s'", key, val, v)
		}
	}
}

func TestLoadFileSave(t *testing.T) {
	f, err := LoadFile("nonexistent")
	if err != nil {
		t.Error("there should not be an error loading nonexistent file")
	}

	if f.Filename != "nonexistent" {
		t.Error("filename should be nonexistent")
	}
}
