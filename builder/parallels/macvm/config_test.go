// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package macvm

import (
	"io/ioutil"
	"os"
	"testing"
)

func testConfig(t *testing.T) map[string]interface{} {
	return map[string]interface{}{
		"ssh_username":     "foo",
		"shutdown_command": "foo",
		"source_path":      "config_test.go",
	}
}

func getTempFile(t *testing.T) *os.File {
	tf, err := ioutil.TempFile("", "packer")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	tf.Close()

	// don't forget to cleanup the file downstream:
	// defer os.Remove(tf.Name())

	return tf
}

func testConfigErr(t *testing.T, warns []string, err error) {
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err == nil {
		t.Fatal("should error")
	}
}

func testConfigOk(t *testing.T, warns []string, err error) {
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err != nil {
		t.Fatalf("bad: %s", err)
	}
}

func TestNewConfig_sourcePath(t *testing.T) {
	// Bad
	c := testConfig(t)
	delete(c, "source_path")
	warns, errs := (&Config{}).Prepare(c)
	testConfigErr(t, warns, errs)

	// Bad
	c = testConfig(t)
	c["source_path"] = "/i/dont/exist"
	warns, errs = (&Config{}).Prepare(c)
	testConfigErr(t, warns, errs)

	// Good
	tf := getTempFile(t)
	defer os.Remove(tf.Name())

	c = testConfig(t)
	c["source_path"] = tf.Name()
	warns, errs = (&Config{}).Prepare(c)
	testConfigOk(t, warns, errs)
}

func TestNewConfig_shutdown_timeout(t *testing.T) {
	cfg := testConfig(t)
	tf := getTempFile(t)
	defer os.Remove(tf.Name())

	// Expect this to fail
	cfg["source_path"] = tf.Name()
	cfg["shutdown_timeout"] = "NaN"
	warns, errs := (&Config{}).Prepare(cfg)
	testConfigErr(t, warns, errs)

	// Passes when given a valid time duration
	cfg["shutdown_timeout"] = "10s"
	warns, errs = (&Config{}).Prepare(cfg)
	testConfigOk(t, warns, errs)
}
