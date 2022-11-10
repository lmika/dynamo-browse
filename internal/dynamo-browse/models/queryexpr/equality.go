package queryexpr

import (
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/lmika/audax/internal/dynamo-browse/models"
	"github.com/lmika/audax/internal/dynamo-browse/models/attrutils"
	"github.com/pkg/errors"
)

func (a *astEqualityOp) evalToIR(info *models.TableInfo) (irAtom, error) {
	leftIR, err := a.Ref.evalToIR(info)
	if err != nil {
		return nil, err
	}

	if a.Op == "" {
		return leftIR, nil
	}

	nameIR, isNameIR := leftIR.(irNamePath)
	if !isNameIR {
		return nil, OperandNotANameError(a.Ref.String())
	}

	rightIR, err := a.Value.evalToIR(info)
	if err != nil {
		return nil, err
	}

	valueIR, isValueIR := rightIR.(irValue)
	if !isValueIR {
		return nil, ValueMustBeLiteralError{}
	}

	switch a.Op {
	case "=":
		return irFieldEq{name: nameIR, value: valueIR}, nil
	case "!=":
		return irFieldNe{name: nameIR, value: valueIR}, nil
	case "^=":
		return irFieldBeginsWith{name: nameIR, value: valueIR}, nil
	}

	return nil, errors.Errorf("unrecognised operator: %v", a.Op)
}

//func (a *astEqualityOp) rightOperandGoValue() (any, error) {
//	if a.Op == "" {
//		return a.Ref.rightOperandGoValue()
//	}
//	return nil, ValueMustBeLiteralError{}
//}
//
//func (a *astEqualityOp) leftOperandName() (string, bool) {
//	return a.Ref.unqualifiedName()
//}

func (a *astEqualityOp) evalItem(item models.Item) (types.AttributeValue, error) {
	left, err := a.Ref.evalItem(item)
	if err != nil {
		return nil, err
	}

	if a.Op == "" {
		return left, nil
	}

	right, err := a.Value.rightOperandDynamoValue()
	if err != nil {
		return nil, err
	}

	switch a.Op {
	case "=":
		cmp, isComparable := attrutils.CompareScalarAttributes(left, right)
		if !isComparable {
			return nil, ValuesNotComparable{Left: left, Right: right}
		}
		return &types.AttributeValueMemberBOOL{Value: cmp == 0}, nil
	case "^=":
		panic("???")
		/*
			rightVal, err := a.Value.rightOperandGoValue()
			if err != nil {
				return nil, err
			}

			strValue, isStrValue := rightVal.(string)
			if !isStrValue {
				return nil, errors.New("operand '^=' must be string")
			}

			leftAsStr, canBeString := attrutils.AttributeToString(left)
			if !canBeString {
				return nil, ValueNotConvertableToString{Val: left}
			}
			return &types.AttributeValueMemberBOOL{Value: strings.HasPrefix(leftAsStr, strValue)}, nil
		*/
	}

	return nil, errors.Errorf("unrecognised operator: %v", a.Op)
}

func (a *astEqualityOp) String() string {
	return a.Ref.String() + a.Op + a.Value.String()
}

type irFieldEq struct {
	name  irNamePath
	value valueIRAtom
}

func (a irFieldEq) canBeExecutedAsQuery(info *models.TableInfo, qci *queryCalcInfo) bool {
	keyName := a.name.keyName()
	if keyName == "" {
		return false
	}

	if keyName == info.Keys.PartitionKey || keyName == info.Keys.SortKey {
		return qci.addKey(info, keyName)
	}

	return false
}

func (a irFieldEq) calcQueryForScan(info *models.TableInfo) (expression.ConditionBuilder, error) {
	nb := a.name.calcName(info)
	vb := a.value.goValue()
	return nb.Equal(expression.Value(vb)), nil
}

func (a irFieldEq) calcQueryForQuery(info *models.TableInfo) (expression.KeyConditionBuilder, error) {
	vb := a.value.goValue()
	return expression.Key(a.name.keyName()).Equal(expression.Value(vb)), nil
}

//func (a irFieldEq) operandFieldName() string {
//	return a.name
//}

type irFieldNe struct {
	name  irNamePath
	value irValue
}

func (a irFieldNe) canBeExecutedAsQuery(info *models.TableInfo, qci *queryCalcInfo) bool {
	return false
}

func (a irFieldNe) calcQueryForScan(info *models.TableInfo) (expression.ConditionBuilder, error) {
	nb := a.name.calcName(info)
	vb := a.value.goValue()
	return nb.NotEqual(expression.Value(vb)), nil
}

func (a irFieldNe) calcQueryForQuery(info *models.TableInfo) (expression.KeyConditionBuilder, error) {
	return expression.KeyConditionBuilder{}, errors.New("cannot use as query")
}

//func (a irFieldNe) operandFieldName() string {
//	return ""
//}

type irFieldBeginsWith struct {
	name  irNamePath
	value irValue
}

func (a irFieldBeginsWith) canBeExecutedAsQuery(info *models.TableInfo, qci *queryCalcInfo) bool {
	keyName := a.name.keyName()
	if keyName == "" {
		return false
	}

	if keyName == info.Keys.SortKey {
		return qci.addKey(info, a.name.name)
	}

	return false
}

func (a irFieldBeginsWith) calcQueryForScan(info *models.TableInfo) (expression.ConditionBuilder, error) {
	nb := a.name.calcName(info)
	vb := a.value.goValue()
	strValue, isStrValue := vb.(string)
	if !isStrValue {
		return expression.ConditionBuilder{}, errors.New("operand '^=' must be string")
	}

	return nb.BeginsWith(strValue), nil
}

func (a irFieldBeginsWith) calcQueryForQuery(info *models.TableInfo) (expression.KeyConditionBuilder, error) {
	vb := a.value.goValue()
	strValue, isStrValue := vb.(string)
	if !isStrValue {
		return expression.KeyConditionBuilder{}, errors.New("operand '^=' must be string")
	}

	return expression.Key(a.name.keyName()).BeginsWith(strValue), nil
}

//func (a irFieldBeginsWith) operandFieldName() string {
//	return a.name
//}

//func getNameAndValue(info *models.TableInfo, name irNamePath, value irValue) (expression.NameBuilder, any, error) {
//nqc, err := name.calcQueryForScan(info)
//if err != nil {
//	return expression.NameBuilder{}, nil, err
//}
//nb := name.calcName(info)

//vqc, err := value.calcQueryForScan(info)
//if err != nil {
//	return nameScanQueryCalc{}, valueScanQueryCalc{}, err
//}
//vb, isVB := vqc.(valueScanQueryCalc)
//if !isVB {
//	return nameScanQueryCalc{}, valueScanQueryCalc{}, errors.New("value is not a value builder")
//}
//	vb := value.goValue()
//
//	return nb, vb, nil
//}
