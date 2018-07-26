package setup

import (
    "io/ioutil"
    "encoding/json"

    "github.com/checkr/flagr/pkg/config"
    "github.com/checkr/flagr/pkg/entity"
    "github.com/checkr/flagr/pkg/repo"
    "github.com/jinzhu/gorm"
    "github.com/sirupsen/logrus"

    yaml "gopkg.in/yaml.v2"
)


// FlagSynchronizer synchronizes YAML config to actual Flagr DB
type FlagSynchronizer struct {
    api *SetupAPI
}

func NewFlagSynchronizer(db *gorm.DB) *FlagSynchronizer {
    if db == nil {
        db = repo.GetDB()
    }
    return &FlagSynchronizer{
        api: &SetupAPI{db: db},
    }
}

// SynchronizeFlags synchronizes flags from the YAML file specified in env.
func (s *FlagSynchronizer) SynchronizeFlags() {
    // Load config from YAML file
    yamlData, err := ioutil.ReadFile(config.Config.YAMLConfigFilePath)
    if err != nil {
        logrus.Fatalf("Unable to read YAML config:", err)
    }
    yamlConfig := &YAMLConfigFile{}
    err = yaml.Unmarshal(yamlData, yamlConfig)
    if err != nil {
        logrus.Fatalf("Unable to unmarshal YAML config: %s.", err)
    }

    // Synchronize each flag
    for flagName, yamlFlag := range yamlConfig.Flags {
        logrus.Infof("Synchronizing flag: %s", flagName)
        flagID := s.synchronizeFlag(flagName, yamlFlag)
        s.api.SaveFlagSnapshot(flagID, "TODO")
    }

    // Delete flags that no longer exist
    // allFlagNames := make([]string, 0, len(yamlConfig.Flags))
    // for key, _ := range yamlConfig.Flags {
    //     allFlagNames = append(allFlagNames, key)
    // }
    s.api.DeleteUnusedFlags(yamlConfig.Flags)
}

func (s *FlagSynchronizer) synchronizeFlag(flagName string, yamlFlag *YAMLFlag) uint {
    flag := s.api.GetOrCreateFlag(flagName)

    // Update flag description and enabled state
    s.api.UpdateFlag(flag, yamlFlag.Description, yamlFlag.Enabled)

    // Create new variants that we need
    variantKeys := make([]string, 0, len(yamlFlag.Variants))
    for _, yamlVariant := range yamlFlag.Variants {
        variantKeys = append(variantKeys, yamlVariant.Key)
    }
    variants := s.api.EnsureVariantsExist(flag.ID, variantKeys)
    variantKeyToID := make(map[string]uint)
    for _, variant := range variants {
        variantKeyToID[variant.Key] = variant.ID
    }

    // Synchronize segments
    s.synchronizeSegments(flag, yamlFlag.Segments, variantKeyToID)

    // Delete un-needed variants, must do this after synchronizing segments
    // because we can't delete a variant until it is no longer being used
    s.api.DeleteUnusedVariants(flag.ID, variantKeyToID)

    return flag.ID
}

func (s *FlagSynchronizer) synchronizeSegments(
    flag *entity.Flag,
    yamlSegments []*YAMLSegment,
    variantKeyToID map[string]uint,
) {
    // Preload flag to get nested objects
    err := flag.Preload(s.api.db)
    if err != nil {
        logrus.Fatalf("Error preloading flag %s: %s", flag.Name, err)
    }
    segments := flag.Segments

    // Synchronization strategy is:
    // While segments exist in DB, modify them to be the same as the YAML.
    // Otherwise create a new segment that matches the YAML. Delete any
    // remaining segments that are no longer used.
    for i, yamlSegment := range yamlSegments {
        var segment *entity.Segment
        if i >= len(segments) {
            // No more segments, create a new one
            segment = s.api.AppendSegment(
                flag.ID, yamlSegment.Description, yamlSegment.Rollout)
        } else {
            segment = &segments[i]
            s.api.UpdateSegment(
                segment, yamlSegment.Description, yamlSegment.Rollout)
        }

        s.synchronizeSegment(flag.ID, segment, yamlSegment, variantKeyToID)
    }

    // Delete any extra segments
    if len(segments) > len(yamlSegments) {
        s.api.DeleteLeftoverSegments(flag.ID, &segments[len(yamlSegments)])
    }
}

func (s *FlagSynchronizer) synchronizeSegment(
    flagID uint,
    segment *entity.Segment,
    yamlSegment *YAMLSegment,
    variantKeyToID map[string]uint,
) {
    // Build constraints and sync them
    constraints := make([]*entity.Constraint, 0, len(yamlSegment.Constraints))
    for _, yamlConstraint := range yamlSegment.Constraints {
        jsonBytes, err := json.Marshal(yamlConstraint.Value)
        if err != nil {
            logrus.Fatalf(
                "Unable to marshal constraint value: %v", yamlConstraint.Value)
        }
        constraints = append(constraints, &entity.Constraint{
            SegmentID: segment.ID,
            Property: yamlConstraint.Property,
            Operator: s.convertOpString(yamlConstraint.Operator),
            Value: string(jsonBytes),
        })
    }
    s.api.EnsureConstraints(segment.ID, constraints)

    // Synchronize distributions
    distributions := make(
        []*entity.Distribution, 0, len(yamlSegment.Distributions))
    var totalPercent uint = 0
    for _, yamlDistribution := range yamlSegment.Distributions {
        distributions = append(distributions, &entity.Distribution{
            SegmentID: segment.ID,
            VariantID: variantKeyToID[yamlDistribution.Key],
            VariantKey: yamlDistribution.Key,
            Percent: yamlDistribution.Percent,
        })
        totalPercent += yamlDistribution.Percent
    }
    if totalPercent != 100 {
        logrus.Fatalf(
            "Distribution percent does not sum to 100 for flag %s", flagID)
    }
    s.api.EnsureDistributions(segment.ID, distributions)

}

func (s *FlagSynchronizer) convertOpString(op string) string {
    switch op {
    case "==":
        return "EQ"
    case "!=":
        return "NEQ"
    case "<":
        return "LT"
    case "<=":
        return "LTE"
    case ">":
        return "GT"
    case ">=":
        return "GTE"
    case "=~":
        return "EREG"
    case "!~":
        return "NEREG"
    case "IN":
        return "IN"
    case "NOT IN":
        return "NOT IN"
    case "CONTAINS":
        return "CONTAINS"
    case "NOT CONTAINS":
        return "NOT CONTAINS"
    default:
        logrus.Fatalf("Unsupported operand %s.", op)
    }
    return ""
}




