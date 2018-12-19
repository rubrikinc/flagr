package setup


// Types for unmarshaling YAML config file

type YAMLConfigFile struct {
    CommonVariants interface{} `yaml:"common_variants"`
    CommonVariantGroups interface{} `yaml:"common_variant_groups"`
    CommonSegments interface{} `yaml:"common_segments"`

    Flags map[string]*YAMLFlag
}

type YAMLFlag struct {
    Description string
    Enabled bool
    Variants []*YAMLVariant
    Segments []*YAMLSegment
}

type YAMLVariant struct {
    Key string
    Attachment map[string]string
}

type YAMLSegment struct {
    Description string
    Rollout uint
    Constraints []*YAMLConstraint
    Distributions []*YAMLDistribution
}

type YAMLConstraint struct {
    Property string
    Operator string
    Value interface{}
}

type YAMLDistribution struct {
    Percent uint
    Key string
}
