package acceptance

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/onsi/gomega/gexec"
	"github.com/pivotal-cf/pivnet-resource/concourse"

	. "github.com/onsi/gomega"
)

func run(command *exec.Cmd, stdinContents []byte) *gexec.Session {
	fmt.Fprintf(os.Stdout, "input: %s\n", stdinContents)

	stdin, err := command.StdinPipe()
	Expect(err).ShouldNot(HaveOccurred())

	session, err := gexec.Start(command, os.Stdout, os.Stdout)
	Expect(err).NotTo(HaveOccurred())

	_, err = io.WriteString(stdin, string(stdinContents))
	Expect(err).ShouldNot(HaveOccurred())

	err = stdin.Close()
	Expect(err).ShouldNot(HaveOccurred())

	return session
}

func metadataValueForKey(metadata []concourse.Metadata, name string) (string, error) {
	for _, i := range metadata {
		if i.Name == name {
			return i.Value, nil
		}
	}
	return "", fmt.Errorf("name not found: %s", name)
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func versionsWithoutFingerprints(versionsWithFingerprints []string) []string {
	versionsWithoutFingerprints := make([]string, len(versionsWithFingerprints))
	for i, v := range versionsWithFingerprints {
		split := strings.Split(v, "#")
		versionsWithoutFingerprints[i] = split[0]
	}
	return versionsWithoutFingerprints
}
