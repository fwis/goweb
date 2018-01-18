package pagination

import (
	"bytes"
	"strconv"

	"github.com/smithfox/goweb/sweb/context"
)

const DEFAULT_PAGE_SIZE = 20

type Pagination struct {
	Paging   bool  //是否分页,如果是false,忽略 其他参数
	DtDraw   int64 //datatable serverSide draw, request and response
	Offset   int64 //request and response
	Total    int64 //总共多少行，可能为空。only response
	PageNum  int   //base on 1, request and response
	PageSize int   //request and response
}

func (m *Pagination) WriteToBufAsJson(buf *bytes.Buffer) {
	if !m.Paging {
		buf.WriteString(`{}`)
		return
	}
	buf.WriteString(`{"Ps":`)
	buf.WriteString(strconv.Itoa(m.PageSize))
	buf.WriteString(`,"Pn":`)
	buf.WriteString(strconv.Itoa(m.PageNum))
	if m.Total > 0 {
		buf.WriteString(`,"Pt":`)
		buf.WriteString(strconv.FormatInt(m.Total, 10))
	}
	if m.DtDraw > 0 {
		buf.WriteString(`,"Draw":`)
		buf.WriteString(strconv.FormatInt(m.DtDraw, 10))
	}
	if m.Offset > 0 {
		buf.WriteString(`,"Offset":`)
		buf.WriteString(strconv.FormatInt(m.Offset, 10))
	}
	buf.WriteString(`}`)
}

//form 参数 pagesize 可选; page为空表示不paging; paging 为 false或0 设置为 paging==false
func NewPagination(context *context.Context, defaultPageSize int) *Pagination {
	pagination := &Pagination{Offset: 0, PageNum: 1, PageSize: defaultPageSize}

	pagination.DtDraw, _ = strconv.ParseInt(context.R.FormValue("dt_draw"), 10, 64)

	pagesize, _ := strconv.Atoi(context.R.FormValue("pagesize"))

	if pagesize > 0 {
		pagination.Paging = true
		pagination.PageSize = pagesize
	}

	if pagination.PageSize <= 0 {
		pagination.PageSize = DEFAULT_PAGE_SIZE
	}

	offset, _ := strconv.ParseInt(context.R.FormValue("offset"), 10, 64)
	if offset > 0 {
		pagination.Paging = true
		pagination.Offset = offset
		pagination.PageNum = 1 + int(float64(pagination.Offset)/float64(pagination.PageSize)+0.1)
	} else {
		pagestr := context.R.FormValue("page")
		var err error
		pagination.PageNum, err = strconv.Atoi(pagestr)
		if err == nil && pagination.PageNum > 0 {
			pagination.Paging = true
		}

		if pagination.PageNum < 1 {
			pagination.PageNum = 1
		}

		pagination.Offset = int64((pagination.PageNum - 1) * pagination.PageSize)
	}
	paging_str := context.R.FormValue("paging")
	if paging_str == "false" || paging_str == "0" {
		pagination.Paging = false
	}
	return pagination
}
