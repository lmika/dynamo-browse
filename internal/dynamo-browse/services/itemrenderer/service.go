package itemrenderer

import (
	"fmt"
	"github.com/lmika/audax/internal/dynamo-browse/models"
	"github.com/lmika/audax/internal/dynamo-browse/models/itemrender"
	"io"
	"text/tabwriter"
)

type Service struct {
	styles styleRenderer
}

func NewService(fileTypeStyle StyleRenderer, metaInfoStyle StyleRenderer) *Service {
	return &Service{
		styles: styleRenderer{
			fileTypeRenderer: fileTypeStyle,
			metaInfoRenderer: metaInfoStyle,
		},
	}
}

func (s *Service) RenderItem(w io.Writer, item models.Item, resultSet *models.ResultSet, plainText bool) {
	styles := s.styles
	if plainText {
		styles = styleRenderer{plainTextStyleRenderer{}, plainTextStyleRenderer{}}
	}

	tabWriter := tabwriter.NewWriter(w, 0, 1, 1, ' ', 0)

	seenColumns := make(map[string]struct{})
	for _, colName := range resultSet.Columns() {
		seenColumns[colName] = struct{}{}
		if r := item.Renderer(colName); r != nil {
			s.renderItem(tabWriter, "", colName, r, styles)
		}
	}
	for k, _ := range item {
		if _, seen := seenColumns[k]; !seen {
			if r := item.Renderer(k); r != nil {
				s.renderItem(tabWriter, "", k, r, styles)
			}
		}
	}
	tabWriter.Flush()
}

func (m *Service) renderItem(w io.Writer, prefix string, name string, r itemrender.Renderer, sr styleRenderer) {
	fmt.Fprintf(w, "%s%v\t%s\t%s%s\n",
		prefix, name, sr.fileTypeRenderer.Render(r.TypeName()), r.StringValue(), sr.metaInfoRenderer.Render(r.MetaInfo()))
	if subitems := r.SubItems(); len(subitems) > 0 {
		for _, si := range subitems {
			m.renderItem(w, prefix+"  ", si.Key, si.Value, sr)
		}
	}
}

type styleRenderer struct {
	fileTypeRenderer StyleRenderer
	metaInfoRenderer StyleRenderer
}
