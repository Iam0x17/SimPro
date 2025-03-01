package support

import (
	"fmt"
	"io"
	"math"
	"strings"
)

type FrameFactoryGen struct {
	w     io.Writer
	defs  []frameDef
	limit int
}

func (g *FrameFactoryGen) Generate(defines GenDefs, w io.Writer) {
	g.w = w
	if g.limit == 0 {
		g.limit = math.MaxInt32
	}
	g.defs = getDefs(g.limit)
	g.generate()
}

func (g *FrameFactoryGen) generate() {
	g.genImports()
	g.genNewFrameFactory()
}

func (g *FrameFactoryGen) genImports() {
	fmt.Fprintf(g.w, "import (\n")
	fmt.Fprintf(g.w, "  \"fmt\"\n")
	fmt.Fprintf(g.w, "  \"SimPro/pkg/go-mysql-server/sql\"\n")
	fmt.Fprintf(g.w, "  \"SimPro/pkg/go-mysql-server/sql/plan\"\n")
	fmt.Fprintf(g.w, "  ast \"github.com/dolthub/vitess/go/vt/sqlparser\"\n")
	fmt.Fprintf(g.w, ")\n\n")
}

func (g *FrameFactoryGen) genNewFrameFactory() {
	fmt.Fprintf(g.w, "func (b *Builder) NewFrame(inScope *scope, f *ast.Frame) sql.WindowFrame {\n")
	fmt.Fprintf(g.w, "  if f == nil {\n")
	fmt.Fprintf(g.w, "    return nil\n")
	fmt.Fprintf(g.w, "  }\n")
	// use manual accessors to init input args
	fmt.Fprintf(g.w, "  isRange := f.Unit == ast.RangeUnit\n")
	fmt.Fprintf(g.w, "  isRows := f.Unit == ast.RowsUnit\n")
	for _, arg := range frameExtents {
		fmt.Fprintf(g.w, "  %s := b.getFrame%s(inScope, f)\n", arg.String(), strings.Title(arg.String()))
	}

	// switch on frame conditionals to select appropriate constructor
	fmt.Fprintf(g.w, "  switch {\n")
	for _, def := range g.defs {
		fmt.Fprintf(g.w, "  case %s:\n", def.CondArgs())
		constructArgs := strings.Builder{}
		i := 0
		for _, a := range def.Args() {
			if a.argType() == "bool" {
				continue
			}
			if i > 0 {
				constructArgs.WriteString(", ")
			}
			constructArgs.WriteString(a.String())
			i++
		}
		fmt.Fprintf(g.w, "    return plan.New%sFrame(%s)\n", def.Name(), constructArgs.String())
	}
	fmt.Fprintf(g.w, "  default:\n")
	fmt.Fprintf(g.w, "    err := fmt.Errorf(\"no matching constructor found for frame: %%v\", f)\n")
	fmt.Fprintf(g.w, "    b.handleErr(err)\n")
	fmt.Fprintf(g.w, "    return nil\n")
	fmt.Fprintf(g.w, "  }\n")
	fmt.Fprintf(g.w, "}\n\n")
}
