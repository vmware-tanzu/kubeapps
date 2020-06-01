import Checkbox from "@material-ui/core/Checkbox";
import FormControlLabel from "@material-ui/core/FormControlLabel";
import Paper from "@material-ui/core/Paper";
import { createStyles, lighten, makeStyles, Theme } from "@material-ui/core/styles";
import Switch from "@material-ui/core/Switch";
import Table from "@material-ui/core/Table";
import TableBody from "@material-ui/core/TableBody";
import TableCell from "@material-ui/core/TableCell";
import TableContainer from "@material-ui/core/TableContainer";
import TableHead from "@material-ui/core/TableHead";
import TableRow from "@material-ui/core/TableRow";
import Toolbar from "@material-ui/core/Toolbar";
import Typography from "@material-ui/core/Typography";
import clsx from "clsx";
import * as React from "react";
import { IChartVersion } from "../../shared/types";

interface IProps {
  versions: IChartVersion[];
  onDeleteFun: any;
  releaseVersion?: string;
}

interface IData {
  version: string;
  appVersion: string;
  description: string;
  createdOn: string;
  isDeployed: string;
}

interface IHeadCell {
  disablePadding: boolean;
  id: keyof IData;
  label: string;
  numeric: boolean;
  minWidth: number;
}

const headCells: IHeadCell[] = [
  { id: "version", numeric: false, disablePadding: true, label: "Chart Version", minWidth: 50 },
  {
    id: "appVersion",
    numeric: false,
    disablePadding: false,
    label: "Application Version",
    minWidth: 75,
  },
  { id: "description", numeric: false, disablePadding: false, label: "Description", minWidth: 170 },
  { id: "createdOn", numeric: false, disablePadding: false, label: "CreatedOn", minWidth: 75 },
  {
    id: "isDeployed",
    numeric: false,
    disablePadding: false,
    label: "Current Deployed",
    minWidth: 50,
  },
];

function EnhancedTableHead() {
  return (
    <TableHead>
      <TableRow>
        <TableCell />
        {headCells.map(headCell => (
          <TableCell
            key={headCell.id}
            align={headCell.numeric ? "right" : "left"}
            padding={headCell.disablePadding ? "none" : "default"}
            style={{ minWidth: headCell.minWidth }}
          >
            {headCell.label}
          </TableCell>
        ))}
      </TableRow>
    </TableHead>
  );
}

const useToolbarStyles = makeStyles((theme: Theme) =>
  createStyles({
    root: {
      paddingLeft: theme.spacing(2),
      paddingRight: theme.spacing(1),
    },
    highlight:
      theme.palette.type === "light"
        ? {
            color: theme.palette.secondary.main,
            backgroundColor: lighten(theme.palette.secondary.light, 0.85),
          }
        : {
            color: theme.palette.text.primary,
            backgroundColor: theme.palette.secondary.dark,
          },
    title: {
      flex: "1 1 100%",
    },
  }),
);

interface IEnhancedTableToolbarProps {
  numSelected: number;
}

const EnhancedTableToolbar = (props: IEnhancedTableToolbarProps) => {
  const classes = useToolbarStyles();
  const { numSelected } = props;

  return (
    <Toolbar
      className={clsx(classes.root, {
        [classes.highlight]: numSelected > 0,
      })}
    >
      {
        <Typography className={classes.title} variant="h6" id="tableTitle" component="div">
          Application Versions
        </Typography>
      }
    </Toolbar>
  );
};

const useStyles = makeStyles((theme: Theme) =>
  createStyles({
    container: {
      maxHeight: 440,
    },
    root: {
      width: "120%",
    },
    paper: {
      width: "120%",
      marginBottom: theme.spacing(5),
    },
    table: {
      minWidth: 750,
    },
    visuallyHidden: {
      border: 0,
      clip: "rect(0 0 0 0)",
      height: 1,
      margin: -1,
      overflow: "hidden",
      padding: 0,
      position: "absolute",
      top: 20,
      width: 1,
    },
  }),
);

export default function DeploymentTableList({ versions, onDeleteFun, releaseVersion }: IProps) {
  const classes = useStyles();
  const [selected, setSelected] = React.useState<string>();
  const [page] = React.useState(0);
  const [dense, setDense] = React.useState(false);
  const [rowsPerPage] = React.useState(5);

  const handleClick = (name: string) => {
    onDeleteFun(name);
    setSelected(name);
  };

  const handleChangeDense = (event: React.ChangeEvent<HTMLInputElement>) => {
    setDense(event.target.checked);
  };

  const isSelected = (name: string) => selected === name;
  const isDefault = (name: string) => !selected && releaseVersion! === name;
  const emptyRows = rowsPerPage - Math.min(rowsPerPage, versions.length - page * rowsPerPage);

  return (
    <div className={classes.root}>
      <Paper className={classes.paper}>
        <EnhancedTableToolbar numSelected={1} />
        <TableContainer className={classes.container}>
          <Table
            className={classes.table}
            aria-labelledby="tableTitle"
            size={dense ? "small" : "medium"}
            stickyHeader={true}
            aria-label="sticky table"
          >
            <EnhancedTableHead />
            <TableBody>
              {versions.map((row, index) => {
                const isItemSelected = isSelected(row.attributes.version);
                const isDefaultSelected = isDefault(row.attributes.version);
                const labelId = `chartversion-${index}`;
                return (
                  <TableRow
                    hover={true}
                    onClick={() => {
                      handleClick(row.attributes.version);
                    }}
                    role="checkbox"
                    aria-checked={isItemSelected}
                    tabIndex={-1}
                    key={row.attributes.version}
                    selected={isItemSelected}
                  >
                    <TableCell padding="checkbox">
                      <Checkbox
                        checked={isItemSelected || isDefaultSelected}
                        inputProps={{ "aria-labelledby": labelId }}
                      />
                    </TableCell>
                    <TableCell
                      component="th"
                      id={labelId}
                      scope="row"
                      padding="none"
                      style={{ minWidth: 50 }}
                    >
                      {row.attributes.version}
                    </TableCell>
                    <TableCell style={{ minWidth: 75 }}>{row.attributes.app_version}</TableCell>
                    <TableCell style={{ minWidth: 170 }}>{row.attributes.description}</TableCell>
                    <TableCell style={{ minWidth: 50 }}>{row.attributes.created}</TableCell>
                    {row.attributes.version === releaseVersion ? (
                      <TableCell size="small" style={{ minWidth: 50 }}>
                        Yes
                      </TableCell>
                    ) : (
                      <TableCell size="small" style={{ minWidth: 50 }} />
                    )}
                  </TableRow>
                );
              })}
              {emptyRows > 0 && (
                <TableRow style={{ height: (dense ? 33 : 53) * emptyRows }}>
                  <TableCell colSpan={6} />
                </TableRow>
              )}
            </TableBody>
          </Table>
        </TableContainer>
      </Paper>
      <FormControlLabel
        control={<Switch checked={dense} onChange={handleChangeDense} />}
        label="Dense padding"
      />
    </div>
  );
}
