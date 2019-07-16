import React, { useContext, useEffect, useState } from 'react'
import clsx from 'clsx'

import Grid from '@material-ui/core/Grid'
import Paper from '@material-ui/core/Paper'
// import Typography from '@material-ui/core/Typography'

import SetTitleContext from 'components/contexts/SetTitleContext'
import { useStyles } from 'components/App/Private/styles'
// import Title from 'components/common/Title'

const Stats = () => {
  const classes = useStyles()
  const [stats, setStats] = useState()
  const setTitle = useContext(SetTitleContext)

  useEffect(() => {
    setTitle('Stats')
  }, [setTitle])

  // useEffect(() => {
  //   statsService.getGlobalStats()
  //     .then((data) => {
  //       setStats(data)
  //     })
  // }, [setTitle])

  const statsMinHeightPaper = clsx(classes.paper, classes.statsMinHeightPaper)

  return (
    <React.Fragment>
      <Grid container spacing={3}>
        <Grid item xs={12} md={6}>
          <Paper className={statsMinHeightPaper}>
            Dashboard
            {/* <AggregatedAgentsStats classes={classes} stats={stats} /> */}
          </Paper>
        </Grid>
        <Grid item xs={12} md={6}>
          <Paper className={statsMinHeightPaper}>
            {/* <SubscribersStats classes={classes} stats={stats} /> */}
          </Paper>
        </Grid>
      </Grid>
    </React.Fragment>
  )
}

export default Stats
