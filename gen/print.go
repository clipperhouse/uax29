// From golang.org/x/text/unicode/rangetable/gen.go

// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Copyright (c) 2009 The Go Authors. All rights reserved.

// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are
// met:

//    * Redistributions of source code must retain the above copyright
// notice, this list of conditions and the following disclaimer.
//    * Redistributions in binary form must reproduce the above
// copyright notice, this list of conditions and the following disclaimer
// in the documentation and/or other materials provided with the
// distribution.
//    * Neither the name of Google Inc. nor the names of its
// contributors may be used to endorse or promote products derived from
// this software without specific prior written permission.

// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
// "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
// LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
// A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
// OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
// SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
// LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
// DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
// THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
// (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
// OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

package main

import (
	"fmt"
	"io"
	"unicode"
)

func print(w io.Writer, rt *unicode.RangeTable) {
	fmt.Fprintln(w, "&unicode.RangeTable{")

	if len(rt.R16) > 0 {
		fmt.Fprintln(w, "\tR16: []unicode.Range16{")
		for _, r := range rt.R16 {
			fmt.Fprintf(w, "\t\t{%#x, %#x, %d},\n", r.Lo, r.Hi, r.Stride)
		}
		fmt.Fprintln(w, "\t},")
	}

	if len(rt.R32) > 0 {
		fmt.Fprintln(w, "\tR32: []unicode.Range32{")
		for _, r := range rt.R32 {
			fmt.Fprintf(w, "\t\t{%#x, %#x, %d},\n", r.Lo, r.Hi, r.Stride)
		}
		fmt.Fprintln(w, "\t},")
	}

	if rt.LatinOffset > 0 {
		fmt.Fprintf(w, "\tLatinOffset: %d,\n", rt.LatinOffset)
	}

	fmt.Fprintf(w, "}\n\n")
}
