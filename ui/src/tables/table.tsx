import Table from "react-bootstrap/Table";

type TableProps = {
  children: React.ReactNode
}

const styles = {
  table: {
    marginLeft: "1%",
    width: "98%",
  },
};

export function AppTable(props: TableProps) {
  return <Table responsive bordered variant="dark" style={styles.table} size="2">
    {props.children}
  </Table>
}