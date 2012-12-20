// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package publicsuffix

import (
	"sort"
	"strings"
	"testing"
)

func TestNodeLabel(t *testing.T) {
	for i, want := range nodeLabels {
		got := nodeLabel(uint32(i))
		if got != want {
			t.Errorf("%d: got %q, want %q", i, got, want)
		}
	}
}

func TestFind(t *testing.T) {
	testCases := []string{
		"",
		"a",
		"a0",
		"aaaa",
		"ao",
		"ap",
		"ar",
		"aro",
		"arp",
		"arpa",
		"arpaa",
		"arpb",
		"az",
		"b",
		"b0",
		"ba",
		"z",
		"zu",
		"zv",
		"zw",
		"zx",
		"zy",
		"zz",
		"zzzz",
	}
	for _, tc := range testCases {
		got := find(tc, 0, numTLD)
		want := notFound
		for i := uint32(0); i < numTLD; i++ {
			if tc == nodeLabel(i) {
				want = i
				break
			}
		}
		if got != want {
			t.Errorf("%q: got %d, want %d", tc, got, want)
		}
	}
}

var publicSuffixTestCases = []struct {
	domain, want string
}{
	// Empty string.
	{"", ""},

	// The .ao rules are:
	// ao
	// ed.ao
	// gv.ao
	// og.ao
	// co.ao
	// pb.ao
	// it.ao
	{"ao", "ao"},
	{"www.ao", "ao"},
	{"pb.ao", "pb.ao"},
	{"www.pb.ao", "pb.ao"},
	{"www.xxx.yyy.zzz.pb.ao", "pb.ao"},

	// The .ar rules are:
	// *.ar
	// !congresodelalengua3.ar
	// !educ.ar
	// !gobiernoelectronico.ar
	// !mecon.ar
	// !nacion.ar
	// !nic.ar
	// !promocion.ar
	// !retina.ar
	// !uba.ar
	// blogspot.com.ar
	{"ar", "ar"},
	{"www.ar", "www.ar"},
	{"nic.ar", "ar"},
	{"www.nic.ar", "ar"},
	{"com.ar", "com.ar"},
	{"www.com.ar", "com.ar"},
	{"blogspot.com.ar", "blogspot.com.ar"},
	{"www.blogspot.com.ar", "blogspot.com.ar"},
	{"www.xxx.yyy.zzz.blogspot.com.ar", "blogspot.com.ar"},
	{"logspot.com.ar", "com.ar"},
	{"zlogspot.com.ar", "com.ar"},
	{"zblogspot.com.ar", "com.ar"},

	// The .arpa rules are:
	// e164.arpa
	// in-addr.arpa
	// ip6.arpa
	// iris.arpa
	// uri.arpa
	// urn.arpa
	{"arpa", "arpa"},
	{"www.arpa", "arpa"},
	{"urn.arpa", "urn.arpa"},
	{"www.urn.arpa", "urn.arpa"},
	{"www.xxx.yyy.zzz.urn.arpa", "urn.arpa"},

	// The relevant {kobe,kyoto}.jp rules are:
	// jp
	// *.kobe.jp
	// !city.kobe.jp
	// kyoto.jp
	// ide.kyoto.jp
	{"jp", "jp"},
	{"kobe.jp", "jp"},
	{"c.kobe.jp", "c.kobe.jp"},
	{"b.c.kobe.jp", "c.kobe.jp"},
	{"a.b.c.kobe.jp", "c.kobe.jp"},
	{"city.kobe.jp", "kobe.jp"},
	{"www.city.kobe.jp", "kobe.jp"},
	{"kyoto.jp", "kyoto.jp"},
	{"test.kyoto.jp", "kyoto.jp"},
	{"ide.kyoto.jp", "ide.kyoto.jp"},
	{"b.ide.kyoto.jp", "ide.kyoto.jp"},
	{"a.b.ide.kyoto.jp", "ide.kyoto.jp"},

	// The .tw rules are:
	// tw
	// edu.tw
	// gov.tw
	// mil.tw
	// com.tw
	// net.tw
	// org.tw
	// idv.tw
	// game.tw
	// ebiz.tw
	// club.tw
	// 網路.tw (xn--zf0ao64a.tw)
	// 組織.tw (xn--uc0atv.tw)
	// 商業.tw (xn--czrw28b.tw)
	// blogspot.tw
	{"tw", "tw"},
	{"aaa.tw", "tw"},
	{"www.aaa.tw", "tw"},
	{"xn--czrw28b.aaa.tw", "tw"},
	{"edu.tw", "edu.tw"},
	{"www.edu.tw", "edu.tw"},
	{"xn--czrw28b.edu.tw", "edu.tw"},
	{"xn--czrw28b.tw", "xn--czrw28b.tw"},
	{"www.xn--czrw28b.tw", "xn--czrw28b.tw"},
	{"xn--uc0atv.xn--czrw28b.tw", "xn--czrw28b.tw"},
	{"xn--kpry57d.tw", "tw"},

	// The .uk rules are:
	// *.uk
	// *.sch.uk
	// !bl.uk
	// !british-library.uk
	// !jet.uk
	// !mod.uk
	// !national-library-scotland.uk
	// !nel.uk
	// !nic.uk
	// !nls.uk
	// !parliament.uk
	// blogspot.co.uk
	{"uk", "uk"},
	{"aaa.uk", "aaa.uk"},
	{"www.aaa.uk", "aaa.uk"},
	{"mod.uk", "uk"},
	{"www.mod.uk", "uk"},
	{"sch.uk", "sch.uk"},
	{"mod.sch.uk", "mod.sch.uk"},
	{"www.sch.uk", "www.sch.uk"},
	{"blogspot.co.uk", "blogspot.co.uk"},
	{"blogspot.nic.uk", "uk"},
	{"blogspot.sch.uk", "blogspot.sch.uk"},

	// The .рф rules are
	// рф (xn--p1ai)
	{"xn--p1ai", "xn--p1ai"},
	{"aaa.xn--p1ai", "xn--p1ai"},
	{"www.xxx.yyy.xn--p1ai", "xn--p1ai"},

	// The .zw rules are:
	// *.zw
	{"zw", "zw"},
	{"www.zw", "www.zw"},
	{"zzz.zw", "zzz.zw"},
	{"www.zzz.zw", "zzz.zw"},
	{"www.xxx.yyy.zzz.zw", "zzz.zw"},

	// There are no .nosuchtld rules.
	{"nosuchtld", "nosuchtld"},
	{"foo.nosuchtld", "nosuchtld"},
	{"bar.foo.nosuchtld", "nosuchtld"},
}

func TestPublicSuffix(t *testing.T) {
	for _, tc := range publicSuffixTestCases {
		got := List.PublicSuffix(tc.domain)
		if got != tc.want {
			t.Errorf("%q: got %q, want %q", tc.domain, got, tc.want)
		}
	}
}

func TestSlowPublicSuffix(t *testing.T) {
	for _, tc := range publicSuffixTestCases {
		got := slowPublicSuffix(tc.domain)
		if got != tc.want {
			t.Errorf("%q: got %q, want %q", tc.domain, got, tc.want)
		}
	}
}

// slowPublicSuffix implements the canonical (but O(number of rules)) public
// suffix algorithm described at http://publicsuffix.org/list/.
//
// 1. Match domain against all rules and take note of the matching ones.
// 2. If no rules match, the prevailing rule is "*".
// 3. If more than one rule matches, the prevailing rule is the one which is an exception rule.
// 4. If there is no matching exception rule, the prevailing rule is the one with the most labels.
// 5. If the prevailing rule is a exception rule, modify it by removing the leftmost label.
// 6. The public suffix is the set of labels from the domain which directly match the labels of the prevailing rule (joined by dots).
// 7. The registered or registrable domain is the public suffix plus one additional label.
//
// This function returns the public suffix, not the registrable domain, and so
// it stops after step 6.
func slowPublicSuffix(domain string) string {
	match := func(rulePart, domainPart string) bool {
		switch rulePart[0] {
		case '*':
			return true
		case '!':
			return rulePart[1:] == domainPart
		}
		return rulePart == domainPart
	}

	domainParts := strings.Split(domain, ".")
	var matchingRules [][]string

loop:
	for _, rule := range rules {
		ruleParts := strings.Split(rule, ".")
		if len(domainParts) < len(ruleParts) {
			continue
		}
		for i := range ruleParts {
			rulePart := ruleParts[len(ruleParts)-1-i]
			domainPart := domainParts[len(domainParts)-1-i]
			if !match(rulePart, domainPart) {
				continue loop
			}
		}
		matchingRules = append(matchingRules, ruleParts)
	}
	if len(matchingRules) == 0 {
		matchingRules = append(matchingRules, []string{"*"})
	} else {
		sort.Sort(byPriority(matchingRules))
	}
	prevailing := matchingRules[0]
	if prevailing[0][0] == '!' {
		prevailing = prevailing[1:]
	}
	if prevailing[0][0] == '*' {
		replaced := domainParts[len(domainParts)-len(prevailing)]
		prevailing = append([]string{replaced}, prevailing[1:]...)
	}
	return strings.Join(prevailing, ".")
}

type byPriority [][]string

func (b byPriority) Len() int      { return len(b) }
func (b byPriority) Swap(i, j int) { b[i], b[j] = b[j], b[i] }
func (b byPriority) Less(i, j int) bool {
	if b[i][0][0] == '!' {
		return true
	}
	if b[j][0][0] == '!' {
		return false
	}
	return len(b[i]) > len(b[j])
}

// TODO(nigeltao): add the "Effective Top Level Domain Plus 1" tests from
// http://mxr.mozilla.org/mozilla-central/source/netwerk/test/unit/data/test_psl.txt
