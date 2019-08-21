import React from 'react'

import Table from '@material-ui/core/Table'
import TableBody from '@material-ui/core/TableBody'
import TableCell from '@material-ui/core/TableCell'
import TableHead from '@material-ui/core/TableHead'
import TableRow from '@material-ui/core/TableRow'

import Title from 'components/common/Title'

const InstanceList = ({ instances }) => (
  <React.Fragment>
    <Title>
      Instances <small>({instances.length})</small>
    </Title>
    <Table size="small">
      <TableHead>
        <TableRow>
          <TableCell>Name</TableCell>
          <TableCell>Plan</TableCell>
          <TableCell>Team</TableCell>
          <TableCell>User</TableCell>
          <TableCell>Status</TableCell>
        </TableRow>
      </TableHead>
      <TableBody>
        {instances.map(instance => (
          <TableRow key={instance.id}>
            <TableCell>{instance.name}</TableCell>
            <TableCell>{instance.plan}</TableCell>
            <TableCell>{instance.team}</TableCell>
            <TableCell>{instance.user}</TableCell>
            <TableCell>{instance.status}</TableCell>
          </TableRow>
        ))}
      </TableBody>
    </Table>
  </React.Fragment>
)

export default InstanceList
