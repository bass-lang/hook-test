package bass_test

import (
	"bytes"
	"errors"
	"testing"

	"github.com/vito/bass/pkg/bass"
	. "github.com/vito/bass/pkg/basstest"
	"github.com/vito/is"
)

type ReaderExample struct {
	Source string
	Result bass.Value
	Err    error
}

func TestReader(t *testing.T) {
	for _, example := range []ReaderExample{
		{
			Source: "_",
			Result: bass.Ignore{},
		},
		{
			Source: "null",
			Result: bass.Null{},
		},
		{
			Source: "false",
			Result: bass.Bool(false),
		},
		{
			Source: "true",
			Result: bass.Bool(true),
		},
		{
			Source: "42",
			Result: bass.Int(42),
		},

		{
			Source: "hello",
			Result: bass.Symbol("hello"),
		},

		{
			Source: ":hello",
			Result: bass.Keyword("hello"),
		},

		{
			Source: ":foo-bar",
			Result: bass.Keyword("foo-bar"),
		},

		{
			Source: `"hello world"`,
			Result: bass.String("hello world"),
		},

		{
			Source: `"hello \"\n\\\t\a\f\r\b\v"`,
			Result: bass.String("hello \"\n\\\t\a\a\r\b\v"),
		},

		{
			Source: `[]`,
			Result: bass.Empty{},
		},
		{
			Source: `[1 true "three"]`,
			Result: bass.Cons{
				A: bass.Int(1),
				D: bass.Cons{
					A: bass.Bool(true),
					D: bass.Cons{
						A: bass.String("three"),
						D: bass.Empty{},
					},
				},
			},
		},

		{
			Source: `{}`,
			Result: bass.Bind{},
		},
		{
			Source: `{:foo 123}`,
			Result: bass.Bind{
				bass.Keyword("foo"), bass.Int(123),
			},
		},
		{
			Source: `{foo 123}`,
			Result: bass.Bind{
				bass.Symbol("foo"), bass.Int(123),
			},
		},
		{
			Source: `{foo}`,
			Result: bass.Bind{bass.Symbol("foo")},
		},

		{
			Source: `()`,
			Result: bass.Empty{},
		},
		{
			Source: `(foo & bar)`,
			Result: bass.Pair{
				A: bass.Symbol("foo"),
				D: bass.Symbol("bar"),
			},
		},
		{
			Source: `(foo 1 & bar)`,
			Result: bass.Pair{
				A: bass.Symbol("foo"),
				D: bass.Pair{
					A: bass.Int(1),
					D: bass.Symbol("bar"),
				},
			},
		},
		{
			Source: `(foo 1 true "three")`,
			Result: bass.Pair{
				A: bass.Symbol("foo"),
				D: bass.NewList(
					bass.Int(1),
					bass.Bool(true),
					bass.String("three"),
				),
			},
		},
		{
			Source: `(foo 1 (two "three"))`,
			Result: bass.Pair{
				A: bass.Symbol("foo"),
				D: bass.NewList(
					bass.Int(1),
					bass.Pair{
						A: bass.Symbol("two"),
						D: bass.NewList(bass.String("three")),
					},
				),
			},
		},

		{
			Source: "./",
			Result: bass.DirPath{
				Path: ".",
			},
		},
		{
			Source: "./foo",
			Result: bass.ExtendPath{
				Parent: bass.DirPath{
					Path: ".",
				},
				Child: bass.FilePath{
					Path: "foo",
				},
			},
		},
		{
			Source: "../",
			Result: bass.DirPath{
				Path: "..",
			},
		},
		{
			Source: "../foo",
			Result: bass.ExtendPath{
				Parent: bass.DirPath{
					Path: "..",
				},
				Child: bass.FilePath{
					Path: "foo",
				},
			},
		},
		{
			Source: "./.foo",
			Result: bass.ExtendPath{
				Parent: bass.DirPath{
					Path: ".",
				},
				Child: bass.FilePath{
					Path: ".foo",
				},
			},
		},
		{
			Source: "./foo/",
			Result: bass.ExtendPath{
				Parent: bass.DirPath{
					Path: ".",
				},
				Child: bass.DirPath{
					Path: "foo",
				},
			},
		},
		{
			Source: ".foo",
			Result: bass.CommandPath{
				Command: "foo",
			},
		},
		{
			Source: "xyz/foo",
			Result: bass.ExtendPath{
				Parent: bass.Symbol("xyz"),
				Child: bass.FilePath{
					Path: "foo",
				},
			},
		},
		{
			Source: "xyz/foo/",
			Result: bass.ExtendPath{
				Parent: bass.Symbol("xyz"),
				Child: bass.DirPath{
					Path: "foo",
				},
			},
		},
		{
			Source: "xyz/foo/bar",
			Result: bass.ExtendPath{
				Parent: bass.ExtendPath{
					Parent: bass.Symbol("xyz"),
					Child: bass.DirPath{
						Path: "foo",
					},
				},
				Child: bass.FilePath{
					Path: "bar",
				},
			},
		},
		{
			Source: "/absolute/path",
			Result: bass.ExtendPath{
				Parent: bass.ExtendPath{
					Parent: bass.DirPath{},
					Child: bass.DirPath{
						Path: "absolute",
					},
				},
				Child: bass.FilePath{
					Path: "path",
				},
			},
		},

		{
			Source: "xyz:foo",
			Result: bass.NewList(
				bass.Keyword("foo"),
				bass.Symbol("xyz"),
			),
		},
		{
			Source: "xyz:foo:bar",
			Result: bass.NewList(
				bass.Keyword("bar"),
				bass.NewList(
					bass.Keyword("foo"),
					bass.Symbol("xyz"),
				),
			),
		},

		{
			Source: "xyz:foo/path",
			Result: bass.ExtendPath{
				Parent: bass.NewList(
					bass.Keyword("foo"),
					bass.Symbol("xyz"),
				),
				Child: bass.FilePath{
					Path: "path",
				},
			},
		},

		{
			Source: `#!/usr/bin/env bass
42`,
			Result: bass.Int(42),
		},

		// quote, syntax-quote, and unquote are not special forms
		{
			Source: `'`,
			Result: bass.Symbol("'"),
		},
		{
			Source: "`",
			Result: bass.Symbol("`"),
		},
		{
			Source: `~`,
			Result: bass.Symbol("~"),
		},
	} {
		example.Run(t)
	}
}

func TestReaderComments(t *testing.T) {
	for _, example := range []ReaderExample{
		{
			Source: `; hello!
_`,
			Result: bass.Annotate{
				Comment: "hello!",
				Value:   bass.Ignore{},
			},
		},
		{
			Source: `;;; hello!
_`,
			Result: bass.Annotate{
				Comment: "hello!",
				Value:   bass.Ignore{},
			},
		},
		{
			Source: `;; ; hello!
_`,
			Result: bass.Annotate{
				Comment: "; hello!",
				Value:   bass.Ignore{},
			},
		},
		{
			Source: `;;;   hello!
_`,
			Result: bass.Annotate{
				Comment: "hello!",
				Value:   bass.Ignore{},
			},
		},
		{
			Source: `; hello!
; multiline!
_`,
			Result: bass.Annotate{
				Comment: "hello! multiline!",
				Value:   bass.Ignore{},
			},
		},
		{
			Source: `; hello!
;
; multi paragraph!
_`,
			Result: bass.Annotate{
				Comment: "hello!\n\nmulti paragraph!",
				Value:   bass.Ignore{},
			},
		},
		{
			Source: `123 ; hello!`,
			Result: bass.Annotate{
				Comment: "hello!",
				Value:   bass.Int(123),
			},
		},
		{
			Source: `; outer
[123 ; hello!
 456 ; goodbye!

 ; inner
 foo
]
`,
			Result: bass.Annotate{
				Comment: "outer",
				Value: bass.NewConsList(
					bass.Annotate{
						Comment: "hello!",
						Value:   bass.Int(123),
					},
					bass.Annotate{
						Comment: "goodbye!",
						Value:   bass.Int(456),
					},
					bass.Annotate{
						Comment: "inner",
						Value:   bass.Symbol("foo"),
					},
				),
			},
		},
	} {
		example.Run(t)
	}
}

func TestReaderMeta(t *testing.T) {
	for _, example := range []ReaderExample{
		{
			Source: `^{:a 1} "since day 1"`,
			Result: bass.Annotate{
				Value: bass.String("since day 1"),
				Meta: &bass.Bind{
					bass.Keyword("a"),
					bass.Int(1),
				},
			},
		},
		{
			Source: `^:b "to thyself"`,
			Result: bass.Annotate{
				Value: bass.String("to thyself"),
				Meta: &bass.Bind{
					bass.Keyword("b"),
					bass.Bool(true),
				},
			},
		},
		{
			Source: `^"hello" "world"`,
			Result: bass.Annotate{
				Value: bass.String("world"),
				Meta: &bass.Bind{
					bass.Keyword("tag"),
					bass.String("hello"),
				},
			},
		},
		{
			Source: `^hello "world"`,
			Result: bass.Annotate{
				Value: bass.String("world"),
				Meta: &bass.Bind{
					bass.Keyword("tag"),
					bass.Symbol("hello"),
				},
			},
		},
		{
			Source: `^[42] "world"`,
			Err:    bass.ErrBadSyntax,
		},
		{
			Source: `; im
			         ^{:a 1} "since day 1"`,
			Result: bass.Annotate{
				Value:   bass.String("since day 1"),
				Comment: "im",
				Meta: &bass.Bind{
					bass.Keyword("a"),
					bass.Int(1),
				},
			},
		},
		{
			Source: `^{:a 1} ^{:b 2} "since week 2"`,
			Result: bass.Annotate{
				Value: bass.String("since week 2"),
				Meta: &bass.Bind{
					bass.Keyword("a"),
					bass.Int(1),
					bass.Keyword("b"),
					bass.Int(2),
				},
			},
		},
	} {
		example.Run(t)
	}
}

func (example ReaderExample) Run(t *testing.T) {
	t.Run(example.Source, func(t *testing.T) {
		is := is.New(t)

		reader := bass.NewReader(bytes.NewBufferString(example.Source))

		form, err := reader.Next()
		if example.Err != nil {
			is.True(errors.Is(err, example.Err))
		} else {
			is.NoErr(err)

			var ann bass.Annotate
			if err := example.Result.Decode(&ann); err != nil {
				var resAnn bass.Annotate
				err := form.Decode(&resAnn)
				is.NoErr(err)
				is.Equal(ann.Comment, resAnn.Comment)
				is.Equal(ann.Meta, resAnn.Meta)
			}

			Equal(t, form, example.Result)
		}
	})
}
