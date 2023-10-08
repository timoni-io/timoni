package db

import (
	"bytes"
	"lib/tlog"
	"lib/utils/maps"

	"github.com/pelletier/go-toml/v2"
)

type GitElementS struct {
	ID          string
	Name        string
	Type        ElementType
	Source      SourceGitS
	Description string
	Favorite    bool   
	UsageCount  int    
	UsageTime   int    
	Error       string 
	FileContent []byte
}

type GitEnvS struct { 
	Name        string
	Source      SourceGitS
	Description string
	Tags        []string
	Teams       []string
	Element     map[string]SourceGitS
	Error       string
	FileContent []byte
}
type GitEnv2S struct { // from TOML file
	Type        string
	Description string
	Tags        []string
	Teams       []string
	Element     map[string]SourceGitS
}

type ElementsInGitRepoCacheS struct {
	LastCommitChecked string
	Elements          []*GitElementS
}

type EnvInGitRepoCacheS struct {
	LastCommitChecked string
	Environments      *maps.SafeMap[string, *GitEnvS]
}

type elementSourceTypeDiscoverS struct {
	Type        ElementType
	Description string
}

type SourceGitS struct {
	RepoName   string `toml:"git-repo-name"`
	BranchName string `toml:"branch"`
	CommitHash string `toml:"commit"`
	CommitTime int64  `toml:"-"`
	FilePath   string `toml:"file-path"`
}

var (
	EnvInGitRepoCache        = *maps.NewSafe[string, EnvInGitRepoCacheS](nil)      // key = gitRepoName/branch
	ElementsInGitRepoCache   = *maps.NewSafe[string, ElementsInGitRepoCacheS](nil) // key = gitRepoName/branch
	ElementsInGitRepoFlatMap = *maps.NewSafe[string, *GitElementS](nil)            // key = gitID
)

func (gitElement *GitElementS) EnvElementGet(elementName, envID string) EnvElementS {

	switch gitElement.Type {
	case ElementSourceTypeConfig:
		element := new(elementConfigS)
		tlog.Error(toml.Unmarshal(gitElement.FileContent, element))
		element.Name = elementName
		element.Type = gitElement.Type
		element.EnvironmentID = envID
		element.SourceGit = gitElement.Source
		return element

	case ElementSourceTypeDomain:
		element := new(elementDomainS)
		tlog.Error(toml.Unmarshal(gitElement.FileContent, element))
		element.Name = elementName
		element.Type = gitElement.Type
		element.EnvironmentID = envID
		element.SourceGit = gitElement.Source
		return element

	case ElementSourceTypeElasticsearch:
		element := new(elementElasticsearchS)
		tlog.Error(toml.Unmarshal(gitElement.FileContent, element))
		element.Name = elementName
		element.Type = gitElement.Type
		element.EnvironmentID = envID
		element.SourceGit = gitElement.Source
		return element

	case ElementSourceTypeMongodb:
		element := new(elementMongodbS)
		tlog.Error(toml.Unmarshal(gitElement.FileContent, element))
		element.Name = elementName
		element.Type = gitElement.Type
		element.EnvironmentID = envID
		element.SourceGit = gitElement.Source
		return element

	case ElementSourceTypePod:
		element := new(elementPodS)
		tlog.Error(toml.Unmarshal(gitElement.FileContent, element))
		element.Name = elementName
		element.Type = gitElement.Type
		element.EnvironmentID = envID
		element.SourceGit = gitElement.Source
		return element
	}

	return nil
}

func (gitEnv *GitEnvS) Validate() (success bool) {

	gitEnv.Error = ""

	if gitEnv.Source.RepoName != "" {

		if gitEnv.Source.BranchName == "" {
			gitEnv.Error = "empty `GitBranchName`"
			return false
		}
		if gitEnv.Source.CommitHash == "" {
			gitEnv.Error = "empty `GitCommitHash`"
			return false
		}
		if gitEnv.Source.FilePath == "" {
			gitEnv.Error = "empty `GitFilePath`"
			return false
		}

		if !reSimpleName1.MatchString(gitEnv.Source.RepoName) {
			gitEnv.Error = "`Source.Git` contains characters that are not allowed: " + gitEnv.Source.RepoName
			return false
		}

	}

	if len(gitEnv.FileContent) == 0 {
		gitEnv.Error = "empty `FileContent`"
		return false
	}

	// (e.g. 'my-name',  or '123-abc.xx'
	if !reSimpleName2.MatchString(gitEnv.Name) {
		gitEnv.Error = "`Source.Name` contains characters that are not allowed: " + gitEnv.Name
		return false
	}

	var err error

	decoder := toml.NewDecoder(bytes.NewReader(gitEnv.FileContent))
	decoder.DisallowUnknownFields()

	newEnv := new(GitEnv2S)

	err = decoder.Decode(newEnv)
	if err != nil {
		errTxt := "Environment file parse fail."
		switch e := err.(type) {
		case *toml.DecodeError:
			errTxt += "\n" + e.String()
		case *toml.StrictMissingError:
			errTxt += "\n" + e.String()
		}
		gitEnv.Error = errTxt
		return false
	}

	gitEnv.Description = newEnv.Description
	gitEnv.Teams = newEnv.Teams
	gitEnv.Tags = newEnv.Tags
	gitEnv.Element = newEnv.Element

	return true
}

// ----------------------

func (gitElement *GitElementS) Validate() (success bool) {

	gitElement.Type = ElementSourceTypeUnknown
	gitElement.Error = ""

	if gitElement.Source.RepoName != "" {

		if gitElement.Source.BranchName == "" {
			gitElement.Error = "empty `GitBranchName`"
			return false
		}
		if gitElement.Source.CommitHash == "" {
			gitElement.Error = "empty `GitCommitHash`"
			return false
		}
		if gitElement.Source.FilePath == "" {
			gitElement.Error = "empty `GitFilePath`"
			return false
		}

		if !reSimpleName1.MatchString(gitElement.Source.RepoName) {
			gitElement.Error = "`Source.Git` contains characters that are not allowed: " + gitElement.Source.RepoName
			return false
		}

	}

	if len(gitElement.FileContent) == 0 {
		gitElement.Error = "empty `FileContent`"
		return false
	}

	// (e.g. 'my-name',  or '123-abc.xx'
	if !reSimpleName2.MatchString(gitElement.Name) {
		gitElement.Error = "`Source.Name` contains characters that are not allowed: " + gitElement.Name
		return false
	}

	var err error

	elementTypeDiscover := new(elementSourceTypeDiscoverS)
	err = toml.Unmarshal(gitElement.FileContent, elementTypeDiscover)
	if err != nil {
		gitElement.Error = "Element " + err.Error() + "\n" + string(gitElement.FileContent)
		return false
	}

	if elementTypeDiscover.Type == "" {
		gitElement.Error = "Element missing field 'type'. Example:\n	type = \"pod\""
		return false
	}

	gitElement.Description = elementTypeDiscover.Description

	decoder := toml.NewDecoder(bytes.NewReader(gitElement.FileContent))
	decoder.DisallowUnknownFields()

	switch elementTypeDiscover.Type {
	// ----------------------------------------------------
	case ElementSourceTypeEnv:
		gitElement.Type = ElementSourceTypeEnv

	case ElementSourceTypePod:

		gitElement.Type = ElementSourceTypePod

		err := decoder.Decode(new(elementPodS))
		if err != nil {
			errTxt := "Element file parse fail."
			switch e := err.(type) {
			case *toml.DecodeError:
				errTxt += "\n" + e.String()
			case *toml.StrictMissingError:
				errTxt += "\n" + e.String()
			}
			gitElement.Error = errTxt
			return false
		}

	case ElementSourceTypeDomain:
		gitElement.Type = ElementSourceTypeDomain

		err := decoder.Decode(new(elementDomainS))
		if err != nil {
			errTxt := "Element file parse fail."
			switch e := err.(type) {
			case *toml.DecodeError:
				errTxt += "\n" + e.String()
			case *toml.StrictMissingError:
				errTxt += "\n" + e.String()
			}
			gitElement.Error = errTxt
			return false
		}

	case ElementSourceTypeConfig:
		gitElement.Type = ElementSourceTypeConfig

		err := decoder.Decode(new(elementConfigS))
		if err != nil {
			errTxt := "Element file parse fail."
			switch e := err.(type) {
			case *toml.StrictMissingError:
				errTxt += "\n" + e.String()
			default:
				errTxt += "\n" + e.Error()
			}
			gitElement.Error = errTxt
			return false
		}

	case ElementSourceTypeElasticsearch:
		gitElement.Type = ElementSourceTypeElasticsearch

		err := decoder.Decode(new(elementElasticsearchS))
		if err != nil {
			errTxt := "Element file parse fail."
			switch e := err.(type) {
			case *toml.DecodeError:
				errTxt += "\n" + e.String()
			case *toml.StrictMissingError:
				errTxt += "\n" + e.String()
			}
			gitElement.Error = errTxt
			return false
		}

	case ElementSourceTypeMongodb:
		gitElement.Type = ElementSourceTypeMongodb

		err := decoder.Decode(new(elementMongodbS))
		if err != nil {
			errTxt := "Element file parse fail."
			switch e := err.(type) {
			case *toml.DecodeError:
				errTxt += "\n" + e.String()
			case *toml.StrictMissingError:
				errTxt += "\n" + e.String()
			}
			gitElement.Error = errTxt
			return false
		}

	default:
		gitElement.Error = "Element unsupported type: `" + string(elementTypeDiscover.Type) + "`"
		return false
	}

	return true
}
