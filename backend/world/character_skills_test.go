package world

import (
	"testing"
)

func TestSkillFlags(test *testing.T) {
	skills := SkillFlags(0)
	skills.Grant(SKILL_AXES)
	skills.Grant(SKILL_NECROMANCY)

	expectedFlag := (SkillFlags(1) << SkillFlags(SKILL_AXES)) | (SkillFlags(1) << SkillFlags(SKILL_NECROMANCY))
	if skills != expectedFlag {
		test.Fatalf("Skills does not meet expected flag. Value is %v, expected %v.", skills, expectedFlag)
	}

	if !skills.Has(SKILL_AXES) {
		test.Fatalf("Skills does not have expected skill AXES")
	}
	if !skills.Has(SKILL_NECROMANCY) {
		test.Fatalf("Skills does not have expected skill NECROMANCY")
	}
	if skills.Has(SKILL_BLUNT) {
		test.Fatalf("Skill has unexpected skill BLUNT")
	}
}
