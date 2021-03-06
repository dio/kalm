import Checkbox from "@material-ui/core/Checkbox";
import TextField, { FilledTextFieldProps } from "@material-ui/core/TextField";
import CheckBoxIcon from "@material-ui/icons/CheckBox";
import CheckBoxOutlineBlankIcon from "@material-ui/icons/CheckBoxOutlineBlank";
import Autocomplete from "@material-ui/lab/Autocomplete";
import { FieldRenderProps } from "react-final-form";
import React from "react";
import { NodeSelectorLabels } from "types/componentTemplate";

const icon = <CheckBoxOutlineBlankIcon fontSize="small" />;
const checkedIcon = <CheckBoxIcon fontSize="small" color="primary" />;

interface Props {
  nodeLabels: string[];
}

export const RenderSelectLabels = ({
  nodeLabels,
  input: { value, onChange },
}: FilledTextFieldProps & FieldRenderProps<NodeSelectorLabels> & Props) => {
  const defaultValue: string[] = [];
  const inputValue = value as NodeSelectorLabels;

  if (inputValue) {
    for (let k in inputValue) {
      defaultValue.push(`${k}:${inputValue[k]}`);
    }
  }

  return (
    <Autocomplete
      multiple
      options={nodeLabels}
      disableCloseOnSelect
      getOptionLabel={(option) => option}
      renderOption={(option, { selected }) => (
        <React.Fragment>
          <Checkbox icon={icon} checkedIcon={checkedIcon} style={{ marginRight: 8 }} checked={selected} />
          {option}
        </React.Fragment>
      )}
      renderInput={(params) => (
        <TextField
          {...params}
          InputLabelProps={{
            shrink: true,
          }}
          variant="outlined"
          label="Node Selector"
          placeholder="Select node labels. Leave blank to schedule on all available nodes."
          helperText="The semantics between labels is AND. A node is a candidate if it match all the labels."
          size={"small"}
        />
      )}
      value={defaultValue}
      onChange={(_, v: any) => {
        const value = v as string[];

        let nodeSelectorLabels: NodeSelectorLabels = {};

        value.forEach((nodeLabel) => {
          const kv = nodeLabel.split(":");

          nodeSelectorLabels[kv[0]] = kv[1];
        });

        onChange(nodeSelectorLabels);
      }}
    />
  );
};
