import React, { useContext, useEffect } from 'react'
import clsx from 'clsx'

import Grid from '@material-ui/core/Grid'
import Paper from '@material-ui/core/Paper'
// import Typography from '@material-ui/core/Typography'

import SetTitleContext from 'components/contexts/SetTitleContext'
import { useStyles } from 'components/App/Private/styles'
// import Title from 'components/common/Title'

const Dashboard = () => {
  const classes = useStyles()
  const setTitle = useContext(SetTitleContext)

  useEffect(() => {
    setTitle('Dashboard')
  }, [setTitle])

  const dashboardMinHeightPaper = clsx(classes.paper, classes.dashboardMinHeightPaper)

  return (
    <React.Fragment>
      <Grid container spacing={3}>
        <Grid item xs={12} md={6}>
          <Paper className={dashboardMinHeightPaper}>
            Dashboard
            TODO
          </Paper>
        </Grid>
        <Grid item xs={12} md={6}>
          <Paper className={dashboardMinHeightPaper}>
            TODO
          </Paper>
        </Grid>
      </Grid>
    </React.Fragment>
  )
}

export default Dashboard
