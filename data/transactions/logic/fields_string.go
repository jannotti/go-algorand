// Code generated by "stringer -type=TxnField,GlobalField,AssetParamsField,AppParamsField,AcctParamsField,AssetHoldingField,OnCompletionConstType,EcdsaCurve,EcGroup,Base64Encoding,JSONRefType,VrfStandard,BlockField -output=fields_string.go"; DO NOT EDIT.

package logic

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[Sender-0]
	_ = x[Fee-1]
	_ = x[FirstValid-2]
	_ = x[FirstValidTime-3]
	_ = x[LastValid-4]
	_ = x[Note-5]
	_ = x[Lease-6]
	_ = x[Receiver-7]
	_ = x[Amount-8]
	_ = x[CloseRemainderTo-9]
	_ = x[VotePK-10]
	_ = x[SelectionPK-11]
	_ = x[VoteFirst-12]
	_ = x[VoteLast-13]
	_ = x[VoteKeyDilution-14]
	_ = x[Type-15]
	_ = x[TypeEnum-16]
	_ = x[XferAsset-17]
	_ = x[AssetAmount-18]
	_ = x[AssetSender-19]
	_ = x[AssetReceiver-20]
	_ = x[AssetCloseTo-21]
	_ = x[GroupIndex-22]
	_ = x[TxID-23]
	_ = x[ApplicationID-24]
	_ = x[OnCompletion-25]
	_ = x[ApplicationArgs-26]
	_ = x[NumAppArgs-27]
	_ = x[Accounts-28]
	_ = x[NumAccounts-29]
	_ = x[ApprovalProgram-30]
	_ = x[ClearStateProgram-31]
	_ = x[RekeyTo-32]
	_ = x[ConfigAsset-33]
	_ = x[ConfigAssetTotal-34]
	_ = x[ConfigAssetDecimals-35]
	_ = x[ConfigAssetDefaultFrozen-36]
	_ = x[ConfigAssetUnitName-37]
	_ = x[ConfigAssetName-38]
	_ = x[ConfigAssetURL-39]
	_ = x[ConfigAssetMetadataHash-40]
	_ = x[ConfigAssetManager-41]
	_ = x[ConfigAssetReserve-42]
	_ = x[ConfigAssetFreeze-43]
	_ = x[ConfigAssetClawback-44]
	_ = x[FreezeAsset-45]
	_ = x[FreezeAssetAccount-46]
	_ = x[FreezeAssetFrozen-47]
	_ = x[Assets-48]
	_ = x[NumAssets-49]
	_ = x[Applications-50]
	_ = x[NumApplications-51]
	_ = x[GlobalNumUint-52]
	_ = x[GlobalNumByteSlice-53]
	_ = x[LocalNumUint-54]
	_ = x[LocalNumByteSlice-55]
	_ = x[ExtraProgramPages-56]
	_ = x[Nonparticipation-57]
	_ = x[Logs-58]
	_ = x[NumLogs-59]
	_ = x[CreatedAssetID-60]
	_ = x[CreatedApplicationID-61]
	_ = x[LastLog-62]
	_ = x[StateProofPK-63]
	_ = x[ApprovalProgramPages-64]
	_ = x[NumApprovalProgramPages-65]
	_ = x[ClearStateProgramPages-66]
	_ = x[NumClearStateProgramPages-67]
	_ = x[invalidTxnField-68]
}

const _TxnField_name = "SenderFeeFirstValidFirstValidTimeLastValidNoteLeaseReceiverAmountCloseRemainderToVotePKSelectionPKVoteFirstVoteLastVoteKeyDilutionTypeTypeEnumXferAssetAssetAmountAssetSenderAssetReceiverAssetCloseToGroupIndexTxIDApplicationIDOnCompletionApplicationArgsNumAppArgsAccountsNumAccountsApprovalProgramClearStateProgramRekeyToConfigAssetConfigAssetTotalConfigAssetDecimalsConfigAssetDefaultFrozenConfigAssetUnitNameConfigAssetNameConfigAssetURLConfigAssetMetadataHashConfigAssetManagerConfigAssetReserveConfigAssetFreezeConfigAssetClawbackFreezeAssetFreezeAssetAccountFreezeAssetFrozenAssetsNumAssetsApplicationsNumApplicationsGlobalNumUintGlobalNumByteSliceLocalNumUintLocalNumByteSliceExtraProgramPagesNonparticipationLogsNumLogsCreatedAssetIDCreatedApplicationIDLastLogStateProofPKApprovalProgramPagesNumApprovalProgramPagesClearStateProgramPagesNumClearStateProgramPagesinvalidTxnField"

var _TxnField_index = [...]uint16{0, 6, 9, 19, 33, 42, 46, 51, 59, 65, 81, 87, 98, 107, 115, 130, 134, 142, 151, 162, 173, 186, 198, 208, 212, 225, 237, 252, 262, 270, 281, 296, 313, 320, 331, 347, 366, 390, 409, 424, 438, 461, 479, 497, 514, 533, 544, 562, 579, 585, 594, 606, 621, 634, 652, 664, 681, 698, 714, 718, 725, 739, 759, 766, 778, 798, 821, 843, 868, 883}

func (i TxnField) String() string {
	if i < 0 || i >= TxnField(len(_TxnField_index)-1) {
		return "TxnField(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _TxnField_name[_TxnField_index[i]:_TxnField_index[i+1]]
}
func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[MinTxnFee-0]
	_ = x[MinBalance-1]
	_ = x[MaxTxnLife-2]
	_ = x[ZeroAddress-3]
	_ = x[GroupSize-4]
	_ = x[LogicSigVersion-5]
	_ = x[Round-6]
	_ = x[LatestTimestamp-7]
	_ = x[CurrentApplicationID-8]
	_ = x[CreatorAddress-9]
	_ = x[CurrentApplicationAddress-10]
	_ = x[GroupID-11]
	_ = x[OpcodeBudget-12]
	_ = x[CallerApplicationID-13]
	_ = x[CallerApplicationAddress-14]
	_ = x[AssetCreateMinBalance-15]
	_ = x[AssetOptInMinBalance-16]
	_ = x[invalidGlobalField-17]
}

const _GlobalField_name = "MinTxnFeeMinBalanceMaxTxnLifeZeroAddressGroupSizeLogicSigVersionRoundLatestTimestampCurrentApplicationIDCreatorAddressCurrentApplicationAddressGroupIDOpcodeBudgetCallerApplicationIDCallerApplicationAddressAssetCreateMinBalanceAssetOptInMinBalanceinvalidGlobalField"

var _GlobalField_index = [...]uint16{0, 9, 19, 29, 40, 49, 64, 69, 84, 104, 118, 143, 150, 162, 181, 205, 226, 246, 264}

func (i GlobalField) String() string {
	if i >= GlobalField(len(_GlobalField_index)-1) {
		return "GlobalField(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _GlobalField_name[_GlobalField_index[i]:_GlobalField_index[i+1]]
}
func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[AssetTotal-0]
	_ = x[AssetDecimals-1]
	_ = x[AssetDefaultFrozen-2]
	_ = x[AssetUnitName-3]
	_ = x[AssetName-4]
	_ = x[AssetURL-5]
	_ = x[AssetMetadataHash-6]
	_ = x[AssetManager-7]
	_ = x[AssetReserve-8]
	_ = x[AssetFreeze-9]
	_ = x[AssetClawback-10]
	_ = x[AssetCreator-11]
	_ = x[invalidAssetParamsField-12]
}

const _AssetParamsField_name = "AssetTotalAssetDecimalsAssetDefaultFrozenAssetUnitNameAssetNameAssetURLAssetMetadataHashAssetManagerAssetReserveAssetFreezeAssetClawbackAssetCreatorinvalidAssetParamsField"

var _AssetParamsField_index = [...]uint8{0, 10, 23, 41, 54, 63, 71, 88, 100, 112, 123, 136, 148, 171}

func (i AssetParamsField) String() string {
	if i < 0 || i >= AssetParamsField(len(_AssetParamsField_index)-1) {
		return "AssetParamsField(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _AssetParamsField_name[_AssetParamsField_index[i]:_AssetParamsField_index[i+1]]
}
func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[AppApprovalProgram-0]
	_ = x[AppClearStateProgram-1]
	_ = x[AppGlobalNumUint-2]
	_ = x[AppGlobalNumByteSlice-3]
	_ = x[AppLocalNumUint-4]
	_ = x[AppLocalNumByteSlice-5]
	_ = x[AppExtraProgramPages-6]
	_ = x[AppCreator-7]
	_ = x[AppAddress-8]
	_ = x[invalidAppParamsField-9]
}

const _AppParamsField_name = "AppApprovalProgramAppClearStateProgramAppGlobalNumUintAppGlobalNumByteSliceAppLocalNumUintAppLocalNumByteSliceAppExtraProgramPagesAppCreatorAppAddressinvalidAppParamsField"

var _AppParamsField_index = [...]uint8{0, 18, 38, 54, 75, 90, 110, 130, 140, 150, 171}

func (i AppParamsField) String() string {
	if i < 0 || i >= AppParamsField(len(_AppParamsField_index)-1) {
		return "AppParamsField(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _AppParamsField_name[_AppParamsField_index[i]:_AppParamsField_index[i+1]]
}
func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[AcctBalance-0]
	_ = x[AcctMinBalance-1]
	_ = x[AcctAuthAddr-2]
	_ = x[AcctTotalNumUint-3]
	_ = x[AcctTotalNumByteSlice-4]
	_ = x[AcctTotalExtraAppPages-5]
	_ = x[AcctTotalAppsCreated-6]
	_ = x[AcctTotalAppsOptedIn-7]
	_ = x[AcctTotalAssetsCreated-8]
	_ = x[AcctTotalAssets-9]
	_ = x[AcctTotalBoxes-10]
	_ = x[AcctTotalBoxBytes-11]
	_ = x[AcctLastProposed-12]
	_ = x[AcctLastHeartbeat-13]
	_ = x[invalidAcctParamsField-14]
}

const _AcctParamsField_name = "AcctBalanceAcctMinBalanceAcctAuthAddrAcctTotalNumUintAcctTotalNumByteSliceAcctTotalExtraAppPagesAcctTotalAppsCreatedAcctTotalAppsOptedInAcctTotalAssetsCreatedAcctTotalAssetsAcctTotalBoxesAcctTotalBoxBytesAcctLastProposedAcctLastHeartbeatinvalidAcctParamsField"

var _AcctParamsField_index = [...]uint16{0, 11, 25, 37, 53, 74, 96, 116, 136, 158, 173, 187, 204, 220, 237, 259}

func (i AcctParamsField) String() string {
	if i < 0 || i >= AcctParamsField(len(_AcctParamsField_index)-1) {
		return "AcctParamsField(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _AcctParamsField_name[_AcctParamsField_index[i]:_AcctParamsField_index[i+1]]
}
func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[AssetBalance-0]
	_ = x[AssetFrozen-1]
	_ = x[invalidAssetHoldingField-2]
}

const _AssetHoldingField_name = "AssetBalanceAssetFrozeninvalidAssetHoldingField"

var _AssetHoldingField_index = [...]uint8{0, 12, 23, 47}

func (i AssetHoldingField) String() string {
	if i < 0 || i >= AssetHoldingField(len(_AssetHoldingField_index)-1) {
		return "AssetHoldingField(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _AssetHoldingField_name[_AssetHoldingField_index[i]:_AssetHoldingField_index[i+1]]
}
func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[NoOp-0]
	_ = x[OptIn-1]
	_ = x[CloseOut-2]
	_ = x[ClearState-3]
	_ = x[UpdateApplication-4]
	_ = x[DeleteApplication-5]
}

const _OnCompletionConstType_name = "NoOpOptInCloseOutClearStateUpdateApplicationDeleteApplication"

var _OnCompletionConstType_index = [...]uint8{0, 4, 9, 17, 27, 44, 61}

func (i OnCompletionConstType) String() string {
	if i >= OnCompletionConstType(len(_OnCompletionConstType_index)-1) {
		return "OnCompletionConstType(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _OnCompletionConstType_name[_OnCompletionConstType_index[i]:_OnCompletionConstType_index[i+1]]
}
func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[Secp256k1-0]
	_ = x[Secp256r1-1]
	_ = x[invalidEcdsaCurve-2]
}

const _EcdsaCurve_name = "Secp256k1Secp256r1invalidEcdsaCurve"

var _EcdsaCurve_index = [...]uint8{0, 9, 18, 35}

func (i EcdsaCurve) String() string {
	if i < 0 || i >= EcdsaCurve(len(_EcdsaCurve_index)-1) {
		return "EcdsaCurve(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _EcdsaCurve_name[_EcdsaCurve_index[i]:_EcdsaCurve_index[i+1]]
}
func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[BN254g1-0]
	_ = x[BN254g2-1]
	_ = x[BLS12_381g1-2]
	_ = x[BLS12_381g2-3]
	_ = x[invalidEcGroup-4]
}

const _EcGroup_name = "BN254g1BN254g2BLS12_381g1BLS12_381g2invalidEcGroup"

var _EcGroup_index = [...]uint8{0, 7, 14, 25, 36, 50}

func (i EcGroup) String() string {
	if i < 0 || i >= EcGroup(len(_EcGroup_index)-1) {
		return "EcGroup(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _EcGroup_name[_EcGroup_index[i]:_EcGroup_index[i+1]]
}
func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[URLEncoding-0]
	_ = x[StdEncoding-1]
	_ = x[invalidBase64Encoding-2]
}

const _Base64Encoding_name = "URLEncodingStdEncodinginvalidBase64Encoding"

var _Base64Encoding_index = [...]uint8{0, 11, 22, 43}

func (i Base64Encoding) String() string {
	if i < 0 || i >= Base64Encoding(len(_Base64Encoding_index)-1) {
		return "Base64Encoding(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _Base64Encoding_name[_Base64Encoding_index[i]:_Base64Encoding_index[i+1]]
}
func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[JSONString-0]
	_ = x[JSONUint64-1]
	_ = x[JSONObject-2]
	_ = x[invalidJSONRefType-3]
}

const _JSONRefType_name = "JSONStringJSONUint64JSONObjectinvalidJSONRefType"

var _JSONRefType_index = [...]uint8{0, 10, 20, 30, 48}

func (i JSONRefType) String() string {
	if i < 0 || i >= JSONRefType(len(_JSONRefType_index)-1) {
		return "JSONRefType(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _JSONRefType_name[_JSONRefType_index[i]:_JSONRefType_index[i+1]]
}
func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[VrfAlgorand-0]
	_ = x[invalidVrfStandard-1]
}

const _VrfStandard_name = "VrfAlgorandinvalidVrfStandard"

var _VrfStandard_index = [...]uint8{0, 11, 29}

func (i VrfStandard) String() string {
	if i < 0 || i >= VrfStandard(len(_VrfStandard_index)-1) {
		return "VrfStandard(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _VrfStandard_name[_VrfStandard_index[i]:_VrfStandard_index[i+1]]
}
func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[BlkSeed-0]
	_ = x[BlkTimestamp-1]
	_ = x[BlkProposer-2]
	_ = x[BlkFeesCollected-3]
	_ = x[invalidBlockField-4]
}

const _BlockField_name = "BlkSeedBlkTimestampBlkProposerBlkFeesCollectedinvalidBlockField"

var _BlockField_index = [...]uint8{0, 7, 19, 30, 46, 63}

func (i BlockField) String() string {
	if i < 0 || i >= BlockField(len(_BlockField_index)-1) {
		return "BlockField(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _BlockField_name[_BlockField_index[i]:_BlockField_index[i+1]]
}
