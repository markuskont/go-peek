package meta

import (
	"fmt"
	"strings"

	"github.com/markuskont/go-sigma-rule-engine/pkg/sigma"
)

const src = `https://raw.githubusercontent.com/mitre/cti/master/enterprise-attack/enterprise-attack.json`

type Tactic struct {
	ID   string
	Name string
}

type Technique struct {
	ID      string
	Name    string
	Tactics []Tactic
}

type MitreAttack struct {
	Technique
	Items      []string
	Techniques []Technique
}

func (m *MitreAttack) Set() *MitreAttack {
	if m.Techniques != nil && len(m.Techniques) > 0 {
		m.Name = fmt.Sprintf("%s: %s", m.Techniques[0].ID, m.Techniques[0].Name)
		m.ID = m.Techniques[0].ID
		m.Items = make([]string, len(m.Techniques))
		for i, t := range m.Techniques {
			m.Items[i] = t.Name
		}
	}
	return m
}

func (m *MitreAttack) ParseSigmaTags(results sigma.Results) *MitreAttack {
	if results == nil || len(results) == 0 {
		return m
	}

	for _, res := range results {
		for _, tag := range res.Tags {
			if strings.HasPrefix(tag, "attack.") {
				if id := strings.Split(tag, "."); len(id) == 2 {
					m.Techniques = append(m.Techniques, Technique{
						ID: strings.ToUpper(id[1]),
					})
				}
			}
		}
	}
	return m
}
