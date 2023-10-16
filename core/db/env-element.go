package db

import (
	"core/config"
	"core/db2"
	"encoding/json"
	"fmt"
	"lib/tlog"
	"lib/utils/conv"
	"lib/utils/maps"
	"lib/utils/set"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	jsonpatch "github.com/evanphx/json-patch/v5"
)

type EnvElementS interface {
	GetName() string
	GetType() ElementType
	GetSource() SourceGitS
	GetStatus() *ElementStatusS
	GetVariablesMap(returnSecrets bool) map[string]ElementVariableS
	GetAutoUpdate() bool
	GetActive() bool
	GetEnvironment() *EnvironmentS
	GetToDelete() bool
	getVersion() int64
	getDBFilePath() string
	check(user *UserS) *tlog.RecordS
	GetScale() *ElementScaleS
	GetVariableError() map[string]ErrorsT

	// cpu, ram
	GetResources() (float64, float64)
	hideSecrets()
	GetUnschedulable() bool
	GetStopped() bool

	Save(user *UserS) *tlog.RecordS
	// Copy fields to el
	Merge(el EnvElementS) *tlog.RecordS

	// Copy secrets to el
	CopySecrets(el EnvElementS) *tlog.RecordS

	// setters - need .Save() after
	setEnvID(envID string)
	generateSecrets(overrideExisting bool)

	SetUnschedulable(bool)
	SetStopped(bool)

	setToDelete(t bool)
	SetAutoUpdate(t bool)
	SetState(state ElementState)
	setVersion(version int64)

	DeleteFromKube() *tlog.RecordS
	KubeApply()

	// pods specific:
	RebuildImage(imageID string, user *UserS) *tlog.RecordS
	RestartAllPods(user *UserS) *tlog.RecordS
	GetImage() *ImageS
	RenderVariables() bool
	SetMetrics(float64, float64)
}

type elementS struct {
	Name          string      `toml:"-"` // uniq name/id in env
	EnvironmentID string      `toml:"-"`
	Type          ElementType `toml:"type"`
	Description   string      `toml:"description"`

	SourceGit  SourceGitS `toml:"-"`
	AutoUpdate bool       `toml:"-"`

	ToDelete      bool `toml:"-"`
	Unschedulable bool `toml:"-"` // when true, Schedule cannot start/stop element
	Stopped       bool `toml:"-"`

	SaveTimestamp    int64  `toml:"-"`
	UserEmail        string `toml:"-"`
	UserInitials     string `toml:"-"`
	CPUUsageAvgCores float64
	RAMUsageAvgMB    float64

	Variables             map[string]*ElementVariableS `toml:"variables"`
	VariablesDependence   *maps.SafeMap[string, bool]  `json:"-" toml:"-"`
	ApplyVariablesOnFiles []string                     `toml:"apply-variables-on-files"`
}

type ElementType string

const (
	ElementSourceTypeUnknown       ElementType = ""
	ElementSourceTypeDomain        ElementType = "domain"
	ElementSourceTypeConfig        ElementType = "config"
	ElementSourceTypePod           ElementType = "pod"
	ElementSourceTypeElasticsearch ElementType = "elasticsearch"
	ElementSourceTypeMongodb       ElementType = "mongodb"
	ElementSourceTypeAction        ElementType = "action"

	ElementSourceTypeEnv ElementType = "env"
)

type ElementVersionS struct {
	SaveTimestamp int64      `json:"SaveTimestamp"`
	UserEmail     string     `json:"UserEmail"`
	UserInitials  string     `json:"UserInitials"`
	SourceGit     SourceGitS `json:"SourceGit"`
	Status        ElementState
	Current       bool
	Previous      bool
}

type ErrorsT map[string]errorMessageT

type ElementVariableS struct {
	Description string  `toml:"description"`
	Secret      bool    `toml:"secret"`
	Validation  string  `toml:"validation"`
	Errors      ErrorsT `toml:"-"` // key=varName  / value=errorMsg
	System      bool    `toml:"-"` // is Read-only

	// FirstValue -> Raw Value
	// mysql://{{config.user}}:{{config.password}}@{{config.host}}
	FirstValue   string `toml:"-"`
	CurrentValue string `toml:"value"`
	// FrontValue - resolved without secrets
	// mysql://myuser:{{config.password}}@myhost
	FrontValue string `toml:"-"`
	// ResolvedValue
	// mysql://myuser:mypassword@myhost
	ResolvedValue string `toml:"-"`
}

type (
	ElementState int
)

const (
	ElementStatusNew         ElementState = 0
	ElementStatusBuilding    ElementState = 5
	ElementStatusDeploying   ElementState = 1
	ElementStatusFailed      ElementState = 2
	ElementStatusReady       ElementState = 3
	ElementStatusTerminating ElementState = 4
	ElementStatusStopped     ElementState = 6
)

var (
	// group 0: {{localvar}} / {{config.refvar}}
	// group 1: localvar / config.refvar
	// group 2: localvar / config
	// group 3: "" / refvar
	regexVariableReference = regexp.MustCompile(`{{(([\w\-]+)\.?([\w\-]+)?)}}`)
	regexVariableName      = regexp.MustCompile(`^[a-zA-Z_]+\w*$`)
	// group 0: int(1, 2) / bool
	// group 1: int / bool
	// group 2: "1, 2" / ""
	regexValidator = regexp.MustCompile(`([a-z]+)(?:\((.*)\))?`)
)

// custom unmarshalers to allow KEY="value" (instead of KEY.value="value")
func (v *ElementVariableS) UnmarshalJSON(data []byte) error {
	type alias ElementVariableS
	return json.Unmarshal(data, (*alias)(v))
}

func (v *ElementVariableS) UnmarshalText(data []byte) error {
	v.FirstValue = string(data)
	return nil
}

// -------------------------------------------------

func ElementFromGitMap(elementName, envID, gitID string) EnvElementS {

	elementInGitRepo := ElementsInGitRepoFlatMap.Get(gitID)
	if elementInGitRepo == nil {
		tlog.Error("element not found in Git repo", tlog.Vars{
			"elementName": elementName,
			"gitID":       gitID,
		})
		return nil
	}

	if elementInGitRepo.Error != "" {
		tlog.Error(elementInGitRepo.Error)
		return nil
	}

	return elementInGitRepo.EnvElementGet(elementName, envID)
}

func ElementFromGit(elementName, envID string, source SourceGitS) EnvElementS {

	gitRepo := GitRepoGetByName(source.RepoName)
	if gitRepo == nil {
		tlog.Error("git-repo not found: " + source.RepoName)
		return nil
	}

	gitRepo.Open()
	defer gitRepo.Unlock()

	if gitRepo.BranchHashMap.Get(source.BranchName) == "" {
		tlog.Error("branch not found: " + source.BranchName)
		return nil
	}

	fileContent := gitRepo.GetFile(source.BranchName, source.CommitHash, source.FilePath)
	if len(fileContent) == 0 {
		tlog.Error("file not found: " + source.FilePath)
		return nil
	}

	source.CommitTime = gitRepo.GetCommit(source.CommitHash).TimeStamp

	gitElement := &GitElementS{
		Name:        filepath.Base(source.FilePath[:len(source.FilePath)-5]),
		Source:      source,
		Type:        ElementSourceTypeUnknown,
		FileContent: fileContent,
	}
	if !gitElement.Validate() {
		tlog.Error(gitElement.Error)
		return nil
	}

	return gitElement.EnvElementGet(elementName, envID)
}

func ElementFromToml(elementName, envID string, src []byte, source SourceGitS) (EnvElementS, *tlog.RecordS) {
	gitElement := &GitElementS{
		Name:        elementName,
		FileContent: src,
		Source:      source,
	}
	if !gitElement.Validate() {

		fmt.Println("----------")
		fmt.Println(gitElement.Error)
		fmt.Println("----------")

		return nil, tlog.Error(gitElement.Error)
	}

	return gitElement.EnvElementGet(elementName, envID), nil
}

// -------------------------------------------------

// k: elementName, v: errorMap
func (element *elementS) GetVariableError() map[string]ErrorsT {

	element.RenderVariables()

	errs := make(map[string]ErrorsT, len(element.Variables))
	for k, v := range element.Variables {
		if len(v.Errors) > 0 {
			errs[k] = v.Errors
		}
	}

	return errs
}

func (element *elementS) GetName() string {
	return element.Name
}

func (element *elementS) GetType() ElementType {
	return element.Type
}

func (element *elementS) GetSource() SourceGitS {
	return element.SourceGit
}

func (element *elementS) GetStatus() *ElementStatusS {

	env := element.GetEnvironment()
	if env == nil {
		return &ElementStatusS{}
	}
	es := env.statuses.Get(element.Name)
	if es == nil {

		if env.statuses == nil {
			env.statuses = maps.NewSafe[string, *ElementStatusS](nil)
		}

		es = &ElementStatusS{}
		env.statuses.Set(element.Name, es)
	}

	return es
}

func (element *elementS) setEnvID(envID string) {
	element.EnvironmentID = envID
}

func (element *elementS) GetAutoUpdate() bool {
	return element.AutoUpdate
}

func (element *elementS) SetUnschedulable(v bool) {
	element.Unschedulable = v
}
func (element *elementS) SetStopped(v bool) {
	element.Stopped = v
}

func (element *elementS) GetUnschedulable() bool {
	return element.Unschedulable
}
func (element *elementS) GetStopped() bool {
	return element.Stopped
}

func (element *elementS) setToDelete(t bool) {
	tlog.Debug("setToDelete", tlog.Vars{
		"envID":       element.EnvironmentID,
		"elementName": element.Name,
	})
	element.SetState(ElementStatusTerminating)
	element.ToDelete = t
}

func (element *elementS) SetAutoUpdate(t bool) {
	if element.SourceGit.RepoName == "" {
		return
	}

	element.AutoUpdate = t
}

func (element *elementS) GetToDelete() bool {
	return element.ToDelete
}

func (element *elementS) GetVariablesMap(returnSecrets bool) map[string]ElementVariableS {
	res := map[string]ElementVariableS{}
	for varName, varData := range element.Variables {
		v := *varData

		if returnSecrets {
			res[varName] = v
			continue
		}

		if varData.Secret {
			// hide secret values
			v.FirstValue = "{{ secret }}"
			v.CurrentValue = "{{ secret }}"
			v.ResolvedValue = "{{ secret }}"
			v.FrontValue = "{{ secret }}"
		} else {
			v.ResolvedValue = v.FrontValue
		}

		res[varName] = v
	}
	return res
}

func (element *elementS) hideSecrets() {
	variables := map[string]*ElementVariableS{}
	for k, v := range element.GetVariablesMap(false) {
		// can't use &v here. Pointer to v is the same for each element,
		// so we need to pass value to some func to make a copy, so it has different address
		variables[k] = conv.Ptr(v)
	}
	element.Variables = variables
}

func (element *elementS) VariablesGet(returnSecrets bool) map[string]string {
	variables := map[string]string{}
	for k, v := range element.GetVariablesMap(returnSecrets) {
		variables[k] = v.ResolvedValue
	}

	if len(element.ApplyVariablesOnFiles) > 0 {
		variables["EP_FIX_ENV"] = strings.Join(element.ApplyVariablesOnFiles, ";")
	}


	return variables
}

func (element *elementS) check(user *UserS) *tlog.RecordS {

	err := ElementNameCheck(element.Name)
	if err != nil {
		return err
	}

	if element.SaveTimestamp <= 0 {
		element.SaveTimestamp = time.Now().Unix()
	}

	if user == nil {
		element.UserEmail = "Timoni"
		element.UserInitials = "T"

	} else {
		element.UserEmail = user.Email
		element.UserInitials = InitialsFromEmail(user.Email)
	}

	// render
	element.RenderVariables()

	return nil
}

func (element *elementS) getDBFilePath() string {
	return filepath.Join(config.DataPath(), "env", element.EnvironmentID, "element", element.Name)
}

func ElementNameCheck(name string) *tlog.RecordS {
	if name == "" {
		return tlog.Error("name is required")
	}
	if !reSimpleName1.MatchString(name) {
		return tlog.Error("name contains characters that are not allowed: " + name)
	}
	if strings.Contains(name, "--") {
		return tlog.Error("name contains `--` that are not allowed")
	}
	if strings.HasPrefix(name, "-") {
		return tlog.Error("name starts with `-` which is not allowed")
	}
	if strings.HasSuffix(name, "-") {
		return tlog.Error("name ends with `-` which is not allowed")
	}
	if strings.HasSuffix(name, "-debug") {
		return tlog.Error("name ends with `-debug` which is not allowed")
	}
	if len(name) > 40 {
		return tlog.Error("name is too long (max 40 chars)")
	}
	if !reSimpleName3.MatchString(string(name[0])) {
		return tlog.Error("element name must start with a letter")

	}

	return nil
}

func (element *elementS) generateSecrets(overrideExistsing bool) {

	for _, varData := range element.Variables {

		if varData.CurrentValue == "" {
			varData.CurrentValue = varData.FirstValue
		}

		if !(varData.Secret && varData.CurrentValue == "") {
			continue
		}

		valueEmpty := varData.ResolvedValue == ""
		override := !valueEmpty && overrideExistsing

		if valueEmpty || override {
			varData.Errors = map[string]errorMessageT{}
			varData.generateSecret()
		}
	}
}

func (element *elementS) RenderVariables() bool {

	element.addSystemVariables()

	element.VariablesDependence = maps.New[string, bool](nil).Safe()

	// first need to generate secrets without overriding them
	// resolveReferences fatals without it
	element.generateSecrets(false)

	// now resolve references
	secrets := set.New[string]()
	for varName, varData := range element.Variables {
		varData.Errors = map[string]errorMessageT{}

		if varData.FirstValue == "" {
			varData.FirstValue = varData.CurrentValue
		}

		element.resolveReferences(
			element.Name+"."+varName,
			varData,
			set.New[string](),
			secrets,
		)
		varData.validateName(varName)
		if len(varData.Errors) > 0 {
			continue
		}
		varData.validateField()
		if len(varData.Errors) > 0 {
			continue
		}
	}

	// ---

	for name, varData := range element.Variables {
		if name == "ELEMENT_GIT_REPO_NAME" {
			// Element can be without repo
			continue
		}
		if varData.ResolvedValue == "" {
			varData.Errors[name] = error_EmptyValue
			return false
		}
	}

	return true
}

func (v *ElementVariableS) generateSecret() {
	v.ResolvedValue = RandString(24)
	// v.FrontValue = "{{ auto-generated }}"
}

func (element *elementS) resolveReferences(varName string, v *ElementVariableS, stack, secrets set.Seter[string]) {
	// TODO: przekazac error do elementu
	if v.Errors == nil {
		v.Errors = map[string]errorMessageT{}
	}

	if v.Secret && v.CurrentValue == "" {
		if v.ResolvedValue == "" {
			tlog.Fatal("Need to call var.generateSecret or element.generareSecrets before resolving references")
		}
		// secret generated
		return
	}

	matches := regexVariableReference.FindAllStringSubmatch(v.CurrentValue, -1)

	if len(matches) == 0 {
		// no references
		v.ResolvedValue = v.CurrentValue
		if v.Secret {
			v.FrontValue = "{{ secret }}"
		} else {
			v.FrontValue = v.CurrentValue
		}
		return
	}

	var env *EnvironmentS
	resolvedValue := v.CurrentValue
	frontValue := v.CurrentValue
	if v.Secret {
		frontValue = "{{ secret }}"
	}

	for _, match := range matches {

		withBrackets := match[0]
		withoutBrackets := match[1]
		var referencedElementName, referencedVariableName string
		var referencedVariablesMap map[string]ElementVariableS

		stack.Add(varName)

		if match[3] == "" {
			// local reference
			referencedElementName = element.Name
			referencedVariableName = match[2]
			referencedVariablesMap = element.GetVariablesMap(true)

		} else {
			// another element reference
			referencedElementName = match[2]
			referencedVariableName = match[3]

			if env == nil {
				env = element.GetEnvironment()
			}

			referencedElement := env.GetElement(referencedElementName)
			if referencedElement == nil {
				tlog.Error("Referenced element not found", tlog.Vars{
					"envID":                  element.EnvironmentID,
					"CurrentValue":           v.CurrentValue,
					"referencedElementName":  referencedElementName,
					"referencedVariableName": referencedVariableName,
				})
				v.Errors[withoutBrackets] = error_ElementNotFound
				element.VariablesDependence.Set(withoutBrackets, false)
				continue
			}

			referencedVariablesMap = referencedElement.GetVariablesMap(true)
		}

		referencedVariable, ok := referencedVariablesMap[referencedVariableName]
		if !ok {
			tlog.Error("Referenced variable not found", tlog.Vars{
				"envID":                  element.EnvironmentID,
				"CurrentValue":           v.CurrentValue,
				"referencedElementName":  referencedElementName,
				"referencedVariableName": referencedVariableName,
			})
			v.Errors[withoutBrackets] = error_VariableNotFound
			element.VariablesDependence.Set(withoutBrackets, false)
			continue
		}
		if len(referencedVariable.Errors) > 0 {
			tlog.Error("Referenced variable has errors", tlog.Vars{
				"envID":                  element.EnvironmentID,
				"CurrentValue":           v.CurrentValue,
				"referencedElementName":  referencedElementName,
				"referencedVariableName": referencedVariableName,
				"errors":                 referencedVariable.Errors,
			})
			v.Errors = referencedVariable.Errors
			element.VariablesDependence.Set(withoutBrackets, false)
			continue
		}

		if referencedVariable.Secret {
			secrets.Add(withoutBrackets)
		}

		if referencedElementName == element.Name && referencedVariable.ResolvedValue == "" {
			// referenced variable not resolved - map order issue
			if stack.Exists(referencedVariableName) {
				// self or cyclic reference
				tlog.Error("Referenced variable has cyclic reference", tlog.Vars{
					"envID":                  element.EnvironmentID,
					"CurrentValue":           v.CurrentValue,
					"referencedElementName":  referencedElementName,
					"referencedVariableName": referencedVariableName,
				})
				v.Errors[withoutBrackets] = error_InvalidReference
				element.VariablesDependence.Set(withoutBrackets, false)
				continue
			}

			// FIXME: this resolves only copy, original still needs to be resolved (doing same job at least twice)
			// can use some cache here
			element.resolveReferences(withoutBrackets, &referencedVariable, stack, secrets)
			if len(referencedVariable.Errors) > 0 {
				element.VariablesDependence.Set(withoutBrackets, false)
				v.Errors = referencedVariable.Errors
				continue
			}
			element.VariablesDependence.Set(withoutBrackets, true)
		}

		resolvedValue = strings.ReplaceAll(
			resolvedValue, withBrackets, referencedVariable.ResolvedValue,
		)

		if !v.Secret && !secrets.Exists(withoutBrackets) {
			frontValue = strings.ReplaceAll(
				frontValue, withBrackets, referencedVariable.FrontValue,
			)
		}
	}

	v.ResolvedValue = resolvedValue
	v.FrontValue = frontValue
}

func (v *ElementVariableS) validateName(name string) {
	if !regexVariableName.MatchString(name) {
		v.Errors[name] = error_InvalidName
	}
}

func (v *ElementVariableS) validateField() {
	v.Validation = strings.TrimSpace(v.Validation)
	if v.Validation == "" {
		return
	}

	if v.Errors == nil {
		v.Errors = map[string]errorMessageT{}
	}

	match := regexValidator.FindStringSubmatch(v.Validation)
	if len(match) == 0 {
		v.Errors[""] = error_InvalidValidator
	}

	fn := match[1]
	args := match[2]

	switch fn {
	case "int":
		args := strings.Split(args, ",")
		if len(args) != 2 {
			v.Errors[""] = error_InvalidValidatorArgs
			return
		}
		for i := range args {
			args[i] = strings.TrimSpace(args[i])
		}

		value, err := strconv.ParseInt(v.ResolvedValue, 10, 64)
		if err != nil {
			v.Errors[""] = error_ValidationFailed
			return
		}
		min, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			v.Errors[""] = error_InvalidValidatorArgs
			return
		}
		max, err := strconv.ParseInt(args[1], 10, 64)
		if err != nil {
			v.Errors[""] = error_InvalidValidatorArgs
			return
		}

		if value < min {
			v.Errors[""] = error_ValidationFailed
			return
		}
		if value > max {
			v.Errors[""] = error_ValidationFailed
			return
		}

	case "bool":
		value := strings.TrimSpace(
			strings.ToLower(
				v.ResolvedValue,
			),
		)
		switch value {
		case "true", "1", "yes", "on", "y", "t":
		case "false", "0", "no", "off", "n", "f":
		default:
			v.Errors[""] = error_ValidationFailed
			return
		}

	case "text":
		args := strings.Split(args, ",")
		if len(args) != 2 {
			v.Errors[""] = error_InvalidValidatorArgs
			return
		}
		for i := range args {
			args[i] = strings.TrimSpace(args[i])
		}

		allowSpaces, err := strconv.ParseBool(args[0])
		if err != nil {
			v.Errors[""] = error_InvalidValidatorArgs
			return
		}
		allowSpecial, err := strconv.ParseBool(args[1])
		if err != nil {
			v.Errors[""] = error_InvalidValidatorArgs
			return
		}

		if !allowSpaces && strings.ContainsAny(v.ResolvedValue, " \n\t\r\f\v\u00A0\u2028\u2029") {
			v.Errors[""] = error_ValidationFailed
			return
		}

		if !allowSpecial && strings.IndexFunc(v.ResolvedValue, func(r rune) bool {
			// return false for digits, letters and space
			// return true for everything else
			//
			// you may need this
			// https://www.ascii-code.com/
			return (r < '0' || r > '9') && (r < 'A' || r > 'Z') && (r < 'a' || r > 'z') && r != ' '
		}) != -1 {
			v.Errors[""] = error_ValidationFailed
			return
		}

	case "oneof":
		if args == "" {
			v.Errors[""] = error_InvalidValidatorArgs
			return
		}
		args := strings.Split(args, ",")
		if len(args) == 0 {
			v.Errors[""] = error_InvalidValidatorArgs
			return
		}
		for _, arg := range args {
			if v.ResolvedValue == strings.TrimSpace(arg) {
				return
			}
		}
		v.Errors[""] = error_ValidationFailed
		return

	case "regex":
		if args == "" {
			v.Errors[""] = error_InvalidValidatorArgs
			return
		}
		re, err := regexp.Compile(args)
		if err != nil {
			v.Errors[""] = error_InvalidValidatorArgs
			return
		}
		if !re.MatchString(v.ResolvedValue) {
			v.Errors[""] = error_ValidationFailed
			return
		}

	case "password":
		args := strings.Split(args, ",")
		if len(args) != 5 {
			v.Errors[""] = error_InvalidValidatorArgs
			return
		}
		for i := range args {
			args[i] = strings.TrimSpace(args[i])
		}

		minLen, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			v.Errors[""] = error_InvalidValidatorArgs
			return
		}
		minLetters, err := strconv.ParseInt(args[1], 10, 64)
		if err != nil {
			v.Errors[""] = error_InvalidValidatorArgs
			return
		}
		upAndLow, err := strconv.ParseBool(args[2])
		if err != nil {
			v.Errors[""] = error_InvalidValidatorArgs
			return
		}
		minDigits, err := strconv.ParseInt(args[3], 10, 64)
		if err != nil {
			v.Errors[""] = error_InvalidValidatorArgs
			return
		}
		minSpecial, err := strconv.ParseInt(args[4], 10, 64)
		if err != nil {
			v.Errors[""] = error_InvalidValidatorArgs
			return
		}

		if x := minLetters + minDigits + minSpecial; x > minLen {
			minLen = x
		}
		if int(minLen) > len(v.ResolvedValue) {
			v.Errors[""] = error_ValidationFailed
			return
		}

		var lower, upper, numbers, special int64
		for _, r := range v.ResolvedValue {
			switch {
			case r >= 'a' && r <= 'z':
				lower++
			case r >= 'A' && r <= 'Z':
				upper++
			case r >= '0' && r <= '9':
				numbers++
			default:
				special++
			}
		}

		if minLetters > 0 && minLetters > lower+upper {
			v.Errors[""] = error_ValidationFailed
			return
		}
		if upAndLow && (lower == 0 || upper == 0) {
			v.Errors[""] = error_ValidationFailed
			return
		}
		if minDigits > 0 && minDigits > numbers {
			v.Errors[""] = error_ValidationFailed
			return
		}
		if minSpecial > 0 && minSpecial > special {
			v.Errors[""] = error_ValidationFailed
			return
		}

	default:
		v.Errors[""] = error_InvalidValidator
	}
}

func (element *elementS) addSystemVariables() {
	if element.Variables == nil {
		element.Variables = map[string]*ElementVariableS{}
	}

	addSystemVariable := func(name, value string) {
		if _, ok := element.Variables[name]; ok {
			return
		}
		element.Variables[name] = &ElementVariableS{
			Errors:        map[string]errorMessageT{},
			FirstValue:    value,
			CurrentValue:  value,
			ResolvedValue: value,
			FrontValue:    value,
			System:        true,
		}
	}

	env := element.GetEnvironment()

	addSystemVariable(
		"NAMESPACE",
		env.ID,
	)
	addSystemVariable(
		"CLUSTER_DOMAIN",
		db2.TheDomain.Name(),
	)
	addSystemVariable(
		"ELEMENT_GIT_REPO_NAME",
		element.SourceGit.RepoName,
	)
	addSystemVariable(
		"ELEMENT_NAME",
		element.Name,
	)
	addSystemVariable(
		"ELEMENT_VERSION",
		fmt.Sprint(element.SaveTimestamp),
	)

}

func (element *elementS) SetState(state ElementState) {
	es := element.GetStatus()
	es.State = state

	if es.Next != nil {
		es.Next.State = state
	}

	es.Save()
}

func (element *elementS) getVersion() int64 {
	return element.SaveTimestamp
}

func (element *elementS) setVersion(version int64) {
	element.SaveTimestamp = version
}

func (element *elementS) CopySecrets(e *elementS) *tlog.RecordS {
	for k, v := range e.Variables {
		if v.Secret && v.CurrentValue == "" {
			if element.Variables[k] == nil {
				tlog.Error("element.Variables[" + k + "] is nil, element=" + element.Name)
				continue
			}
			*v = *element.Variables[k]
		}
	}
	return nil
}

func (element *elementS) Merge(e *elementS) *tlog.RecordS {

	if element == nil {
		tlog.Error("element is nil")
		return nil
	}

	e.AutoUpdate = element.AutoUpdate
	e.ToDelete = element.ToDelete
	e.Unschedulable = element.Unschedulable
	e.Stopped = element.Stopped
	e.UserEmail = element.UserEmail
	e.UserInitials = element.UserInitials
	for k, v := range e.Variables {
		if v.Secret && v.CurrentValue == "" {
			if element.Variables[k] == nil {
				tlog.Error("element.Variables[" + k + "] is nil, element=" + element.Name)
				continue
			}
			*v = *element.Variables[k]
		}
	}
	return nil
}

func CreatePatch(old, new EnvElementS) ([]byte, error) {
	// mask fields
	old.setVersion(0)
	new.setVersion(0)

	a, err := json.Marshal(old)
	if err != nil {
		return nil, err
	}
	b, err := json.Marshal(new)
	if err != nil {
		return nil, err
	}

	return jsonpatch.CreateMergePatch(a, b)
}

func ApplyPatch(in EnvElementS, patch ...byte) (EnvElementS, *tlog.RecordS) {
	data, err := json.Marshal(in)
	if err != nil {
		return nil, tlog.Error(err)
	}

	if len(patch) == 0 {
		path := in.getDBFilePath()
		os.MkdirAll(path, 0755)
		patch, err = os.ReadFile(filepath.Join(path, "0-patch.json"))
		if err != nil {
			return in, nil
		}
	}

	data, err = jsonpatch.MergePatch(data, patch)
	if err != nil {
		return nil, tlog.Error(err)
	}

	out := reflect.New(reflect.Indirect(reflect.ValueOf(in)).Type()).Interface().(EnvElementS)
	err = json.Unmarshal(data, out)
	if err != nil {
		return nil, tlog.Error(err)
	}
	return out, nil
}

func ElementUpdate(elOld, elNew EnvElementS, user *UserS, byWebUI bool) *tlog.RecordS {
	if elNew == nil {
		return tlog.Error("elNew is nil")
	}

	tlog.Info("element update "+elOld.GetName(), tlog.Vars{
		"env":     elOld.GetEnvironment().ID,
		"element": elOld.GetName(),
		"event":   "true",
	})

	elNew.SetState(ElementStatusDeploying)

	es := elNew.GetSource()
	elNew.RenderVariables()

	if elOld.GetSource().CommitHash != es.CommitHash {
		elNew.GetStatus().Next = &ElementNextS{
			SourceGit:   es,
			StepCurrent: 1,
			StepCount:   7,
			State:       ElementStatusNew,
			Message:     "Git checkout",
		}
	}

	elNew.GetStatus().NewerVersion = es.CommitHash != GitRepoGetByName(es.RepoName).GetLastCommit(es.BranchName)

	if err := elOld.Merge(elNew); err != nil {
		return err
	}

	if !byWebUI {
		// git version change
		// skip creating patch
		return elNew.Save(user)
	}
	// front element update
	// create patch and save it

	patch, err := CreatePatch(elOld, elNew)
	if err != nil {
		return tlog.Error(err.Error())
	}

	return savePatch(elNew, patch, user)
}

type ElementHistory struct {
	Data  json.RawMessage
	Patch json.RawMessage
}

func SaveHistory(element EnvElementS) *tlog.RecordS {
	path := element.getDBFilePath()
	os.MkdirAll(path, 0755)

	fiName := filepath.Join(path, "0-current.json")
	if _, err := os.Stat(fiName); os.IsNotExist(err) {
		return tlog.Error("current version not saved. cannot save history", tlog.Vars{
			"envID":       element.GetEnvironment().ID,
			"elementName": element.GetName(),
			"fiName":      fiName,
		})
	}

	prevData, err := os.ReadFile(fiName)
	if err != nil {
		return tlog.Error(err.Error())
	}

	patch, _ := os.ReadFile(filepath.Join(path, "0-patch.json"))
	if len(patch) == 0 {
		patch = []byte("{}")
	}

	// check if element has changed
	a, _ := jsonpatch.MergePatch(prevData, patch)
	b, _ := json.Marshal(element)

	if string(a) == string(b) {
		// skip history creation
		return nil
	}

	history, err := json.MarshalIndent(ElementHistory{
		Data:  prevData,
		Patch: patch,
	}, "", "\t")
	if err != nil {
		return tlog.Error(err.Error())
	}

	err = os.WriteFile(filepath.Join(path, fmt.Sprintf("%d.json", element.getVersion())), history, 0644)
	if err != nil {
		return tlog.Error(err.Error())
	}
	return nil
}

func savePatch(element EnvElementS, patch []byte, user *UserS) *tlog.RecordS {
	{
		err := element.check(user)
		if err != nil {
			return err
		}
	}
	SaveHistory(element)

	// Update patch
	path := element.getDBFilePath()
	os.MkdirAll(path, 0755)

	err := os.WriteFile(filepath.Join(path, "0-patch.json"), patch, 0644)
	if err != nil {
		return tlog.Error(err.Error())
	}

	// update cache
	env := element.GetEnvironment()
	if env == nil {
		return tlog.Error("env == nil")
	}

	element, _ = ApplyPatch(element, patch...)
	ElementMap.Set(fmt.Sprintf("%s/%s", env.ID, element.GetName()), element)
	return nil
}

func elementSave(element EnvElementS, user *UserS) *tlog.RecordS {
	SaveHistory(element)

	// update element
	data, err := json.MarshalIndent(element, "", "\t")
	if err != nil {
		return tlog.Error(err.Error())
	}

	path := element.getDBFilePath()
	os.MkdirAll(path, 0755)

	err = os.WriteFile(filepath.Join(path, "0-current.json"), data, 0644)
	if err != nil {
		return tlog.Error(err.Error())
	}

	// update cache
	env := element.GetEnvironment()
	if env == nil {
		return tlog.Error("env == nil")
	}

	if !element.GetToDelete() {
		element, _ = ApplyPatch(element)
	}
	ElementMap.Set(fmt.Sprintf("%s/%s", env.ID, element.GetName()), element)
	return nil
}

func (element *elementS) GetActive() bool {

	es := element.GetStatus()

	if element.ToDelete {
		if es.State != ElementStatusTerminating {
			element.SetState(ElementStatusTerminating)
		}
		return false
	}

	if element.Stopped {
		if es.State != ElementStatusStopped {
			element.SetState(ElementStatusStopped)
		}
		return false
	} else {
		if es.State == ElementStatusStopped {
			element.SetState(ElementStatusNew)
		}
	}

	if es.State == ElementStatusTerminating {
		return false
	}

	if es.State == ElementStatusStopped {
		return false
	}

	return true
}

func (element *elementS) GetEnvironment() *EnvironmentS {
	return EnvironmentMap.Get(element.EnvironmentID)
}

func stepSuccess(msg string) CheckStepResultS {

	return CheckStepResultS{
		Status: CheckStepResultSuccess,
		Msg:    msg,
	}
}

func stepFail(msg string) CheckStepResultS {
	return CheckStepResultS{
		Status: CheckStepResultFail,
		Msg:    msg,
	}
}

func stepInProgress(msg string) CheckStepResultS {
	return CheckStepResultS{
		Status: CheckStepResultInProgress,
		Msg:    msg,
	}
}

func GetElementPod(element EnvElementS) (*elementPodS, bool) {
	e, ok := element.(*elementPodS)
	return e, ok
}

func GetElementDomain(element EnvElementS) (*elementDomainS, bool) {
	e, ok := element.(*elementDomainS)
	return e, ok
}

func GetElementAction(element EnvElementS) (*elementActionS, bool) {
	e, ok := element.(*elementActionS)
	return e, ok
}

func (e *elementS) GetResources() (cpu, ram float64) {
	return e.CPUUsageAvgCores, e.RAMUsageAvgMB
}

func (e *elementS) SetMetrics(cpu, ram float64) {
	e.CPUUsageAvgCores = cpu
	e.RAMUsageAvgMB = ram
}
