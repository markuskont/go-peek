package sigma

func isKeyword(s string) bool {

}

/*
var (
	matchOrder = []token{
		tokenOne,
		tokenAll,
		tokenAgg,
		tokenNear,
		tokenBy,
		tokenEq,
		tokenLt,
		tokenLte,
		tokenGt,
		tokenGte,
		tokenPipe,
		tokenAnd,
		tokenOr,
		tokenNot,
		tokenId,
		tokenLpar,
		tokenRpar,
		tokenEof,
	}
	tokenDefs = func() []tokenDef {
		defs := make([]tokenDef, len(matchOrder))
		for i, t := range matchOrder {
			defs[i] = tokenDef{
				token:  t,
				Regexp: t.Pattern(),
			}
		}
		return defs
	}()
)

type conditionToken struct {
	position []int
	token    token
}

func parseCondition(condition string) []conditionToken {
	tokens := []conditionToken{}
	return tokens
}

type Condtition int

const (
	CondNone Condtition = iota
	CondAnd
	CondOr
	CondNot
	CondNull
)

type token int

const (
	tokenInvalid token = iota
	tokenAnd
	tokenOr
	tokenNot
	tokenId
	tokenLpar
	tokenRpar
	tokenPipe
	tokenOne
	tokenAll
	tokenAgg
	tokenEq
	tokenLt
	tokenLte
	tokenGt
	tokenGte
	tokenBy
	tokenNear
	tokenEof
)

func (t token) String() string {
	switch t {
	case tokenAnd:
		return "AND"
	case tokenOr:
		return "OR"
	case tokenNot:
		return "NOT"
	case tokenId:
		return "ID"
	case tokenLpar:
		return "LPAR"
	case tokenRpar:
		return "RPAR"
	case tokenPipe:
		return "PIPE"
	case tokenOne:
		return "ONE"
	case tokenAll:
		return "ALL"
	case tokenAgg:
		return "AGG"
	case tokenEq:
		return "EQ"
	case tokenLt:
		return "LT"
	case tokenLte:
		return "LTE"
	case tokenGt:
		return "GT"
	case tokenGte:
		return "GTE"
	case tokenBy:
		return "BY"
	case tokenNear:
		return "NEAR"
	case tokenEof:
		return "EOF"
	default:
		return "INVALID"
	}
}

func (t token) Pattern() *regexp.Regexp {
	switch t {
	case tokenAnd:
		return regexp.MustCompile(`(?i)and`)
	case tokenOr:
		return regexp.MustCompile(`(?i)or`)
	case tokenNot:
		return regexp.MustCompile(`(?i)not`)
	case tokenId:
		return regexp.MustCompile(`^[\w*]+$`)
	case tokenLpar:
		return regexp.MustCompile(`\(`)
	case tokenRpar:
		return regexp.MustCompile(`\)`)
	case tokenPipe:
		return regexp.MustCompile(`\|`)
	case tokenOne:
		return regexp.MustCompile(`1 of`)
	case tokenAll:
		return regexp.MustCompile(`all of`)
	case tokenAgg:
		return regexp.MustCompile(`(?i)count|min|max|avg|sum`)
	case tokenEq:
		return regexp.MustCompile(`==`)
	case tokenLt:
		return regexp.MustCompile(`<`)
	case tokenLte:
		return regexp.MustCompile(`<=`)
	case tokenGt:
		return regexp.MustCompile(`>`)
	case tokenGte:
		return regexp.MustCompile(`>=`)
	case tokenBy:
		return regexp.MustCompile(`(?i)by`)
	case tokenNear:
		return regexp.MustCompile(`(?i)near`)
	default:
		return nil
	}
}

type tokenDef struct {
	token
	*regexp.Regexp
}
*/
