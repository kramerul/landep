package landep

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
)

func defaultConflictSolver(path string, j1 json.RawMessage, j2 json.RawMessage) (json.RawMessage, error) {
	return nil, fmt.Errorf("Incompatible jsons at %s: '%s' '%s'", path, string(j1), string(j2))
}

func MaximumConflictSolver(path string, j1 json.RawMessage, j2 json.RawMessage) (json.RawMessage, error) {
	var i1 int
	var i2 int
	err := json.Unmarshal(j1, &i1)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(j2, &i2)
	if err != nil {
		return nil, err
	}
	if i1 < i2 {
		i1 = i2
	}
	return json.Marshal(&i1)
}

func JsonMerge(jsons []json.RawMessage) (json.RawMessage, error) {
	return jsonMerge(jsons, "", defaultConflictSolver)
}

func JsonMergeWithConflictSolver(jsons []json.RawMessage, conflictSolver func(path string, j1 json.RawMessage, j2 json.RawMessage) (json.RawMessage, error)) (json.RawMessage, error) {
	return jsonMerge(jsons, "", conflictSolver)
}

func jsonMerge(jsons []json.RawMessage, path string, conflictSolver func(path string, j1 json.RawMessage, j2 json.RawMessage) (json.RawMessage, error)) (json.RawMessage, error) {
	if len(jsons) == 0 {
		return nil, nil
	}
	if len(jsons) == 1 {
		return jsons[0], nil
	}
	ok := make([]bool, len(jsons))
	maps := make([]map[string]json.RawMessage, len(jsons))
	var err error
	for i, j := range jsons {
		maps[i], ok[i], err = JsonMappify(j)
		if err != nil {
			return nil, err
		}
	}
	for i := 1; i < len(ok); i++ {
		if ok[i-1] != ok[i] {
			return nil, fmt.Errorf("Incompatible jsons types: '%s' '%s'", string(jsons[i-1]), string(jsons[i]))
		}
	}
	if ok[0] {
		values := map[string][]json.RawMessage{}
		for _, m := range maps {
			for k, v := range m {
				values[k] = append(values[k], v)
			}
		}
		result := map[string]json.RawMessage{}
		for k, v := range values {
			result[k], err = jsonMerge(v, path+"."+k, conflictSolver)
			if err != nil {
				return nil, err
			}
		}
		return json.Marshal(result)
	}
	result := jsons[0]
	for i := 1; i < len(jsons); i++ {
		if bytes.Compare(result, jsons[i]) != 0 {
			result, err = conflictSolver(path, result, jsons[i])
			if err != nil {
				return nil, err
			}
		}
	}
	return result, nil
}

var objectRegexp = regexp.MustCompile("^\\s*\\{")

func JsonMappify(i json.RawMessage) (map[string]json.RawMessage, bool, error) {
	if objectRegexp.Match(i) {
		m := map[string]json.RawMessage{}
		err := json.Unmarshal(i, &m)
		if err != nil {
			return nil, false, err
		}
		return m, true, nil
	}
	return nil, false, nil
}
