package setup

import (
    "github.com/checkr/flagr/pkg/entity"

    "github.com/jinzhu/gorm"
    "github.com/sirupsen/logrus"

)

// Helper for making DB calls, and logging fatal if an error was found
func fatalIfDBErrors(db *gorm.DB) {
    errs := db.GetErrors()
    if len(errs) > 0 {
        logrus.Fatalf("Errors while making DB call: %v", errs)
    }
}

// SetupAPI contains helper functions for manipualting the DB when
// synchronizing flags during setup.
type SetupAPI struct {
    db *gorm.DB
}

func (api *SetupAPI) SaveFlagSnapshot(flagID uint, updatedBy string) {
    entity.SaveFlagSnapshot(api.db, flagID, updatedBy)
}

func (api *SetupAPI) DeleteUnusedFlags(flagNames map[string]*YAMLFlag) {
    count, err := entity.NewFlagQuerySet(api.db).Count()
    if err != nil {
        logrus.Fatalf("Error while counting flags: %s", err)
    }
    allFlags := make([]entity.Flag, 0, count)
    err = entity.NewFlagQuerySet(api.db).All(&allFlags)
    if err != nil {
        logrus.Fatalf("Error while listing all flags: %s", err)
    }

    for _, flag := range allFlags {
        // Delete flag if its not in the given flagNames
        if _, ok := flagNames[flag.Name]; !ok {
            logrus.Infof("Deleting flag: %s", flag.Name)
            if err := flag.Delete(api.db); err != nil {
                logrus.Fatalf("Error while deleting flags: %s", err)
            }
        }
    }
}

func (api *SetupAPI) GetOrCreateFlag(flagName string) *entity.Flag {
    flag := &entity.Flag{}
    fatalIfDBErrors(
        api.db.FirstOrCreate(flag, entity.Flag{Name: flagName}),
    )
    return flag
}

func (api *SetupAPI) UpdateFlag(flag *entity.Flag, description string, enabled bool) {
    flag.Description = description
    flag.Enabled = enabled
    fatalIfDBErrors(api.db.Save(flag))
}

func (api *SetupAPI) EnsureVariantsExist(flagID uint, keys []string) []*entity.Variant {
    result := make([]*entity.Variant, 0, len(keys))
    for _, key := range keys {
        variant := &entity.Variant{}
        fatalIfDBErrors(
            api.db.FirstOrCreate(
                variant,
                entity.Variant{
                    FlagID: flagID,
                    Key: key,
                    Attachment: entity.Attachment{},
                },
            ),
        )
        result = append(result, variant)
    }
    return result
}

func (api *SetupAPI) DeleteUnusedVariants(flagID uint, keys map[string]uint) {
    count, err := entity.NewVariantQuerySet(api.db).Count()
    if err != nil {
        logrus.Fatalf("Error while counting variants: %s", err)
    }
    allVariants := make([]entity.Variant, 0, count)
    err = entity.NewVariantQuerySet(api.db).All(&allVariants)
    if err != nil {
        logrus.Fatalf("Error while listing all variants: %s", err)
    }

    for _, variant := range allVariants {
        if _, ok := keys[variant.Key]; !ok {
            if err := variant.Delete(api.db); err != nil {
                logrus.Fatalf("Error while deleting variants: %s", err)
            }
        }
    }
}

func (api *SetupAPI) AppendSegment(
    flagID uint,
    description string,
    rollout uint,
) *entity.Segment {
    segment := &entity.Segment{
        FlagID: flagID,
        Description: description,
        RolloutPercent: rollout,
        Rank: entity.SegmentDefaultRank,
    }
    if err := segment.Create(api.db); err != nil {
        logrus.Fatalf("Error creating segment: %s", err)
    }
    return segment
}

func (api *SetupAPI) DeleteLeftoverSegments(flagID uint, segment *entity.Segment) {
    // Delete all segments that are ordered after the given one
    q := entity.NewSegmentQuerySet(api.db)
    if err := q.FlagIDEq(flagID).RankGt(segment.Rank).Delete(); err != nil {
        logrus.Fatalf("Error deleting segments: %s", err)
    }

    // Also delete elements that are tied in rank but have ID gte
    q = entity.NewSegmentQuerySet(api.db)
    q = q.FlagIDEq(flagID).RankEq(segment.Rank).IDGte(segment.ID)
    if err := q.Delete(); err != nil {
        logrus.Fatalf("Error deleting segments: %s", err)
    }
}

func (api *SetupAPI) UpdateSegment(
    segment *entity.Segment,
    description string,
    rollout uint,
) {
    segment.Description = description
    segment.RolloutPercent = rollout
    fatalIfDBErrors(api.db.Save(segment))
}

func (api *SetupAPI) EnsureConstraints(
    segmentID uint,
    constraints []*entity.Constraint,
) {
    // Delete every constraint and re-create them
    q := entity.NewConstraintQuerySet(api.db).SegmentIDEq(segmentID)
    if err := q.Delete(); err != nil {
        logrus.Fatalf("Failed to delete constraints: %s", err)
    }

    for _, constraint := range constraints {
        if err := constraint.Create(api.db); err != nil {
            logrus.Fatalf("Unable to create constraint: %s", err)
        }
    }
}

func (api *SetupAPI) EnsureDistributions(
    segmentID uint,
    distributions []*entity.Distribution,
) {
    // Delete every distribution and re-create them
    q := entity.NewDistributionQuerySet(api.db).SegmentIDEq(segmentID)
    if err := q.Delete(); err != nil {
        logrus.Fatalf("Failed to delete distributions: %s", err)
    }

    for _, distribution := range distributions {
        if err := distribution.Create(api.db); err != nil {
            logrus.Fatalf("Unable to create distribution: %s", err)
        }
    }
}
